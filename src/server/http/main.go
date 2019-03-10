package main

import (
	"encoding/gob"
	"errors"
	"storage"
	"strconv"
	"time"
	"weibo"

	"github.com/gomodule/redigo/redis"

	"github.com/boj/redistore"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/weibo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	redisClient, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	userRepo := storage.NewUserRepository(db, redisClient)
	timelineRepo := storage.NewTimeLineRepository(db)
	weiboReppo := storage.NewWeiboRepository(db, redisClient)
	service := weibo.NewService(userRepo, weiboReppo, timelineRepo)

	gob.Register(new(weibo.User))
	// var store = sessions.NewCookieStore([]byte("test"))
	store, err := redistore.NewRediStore(10, "tcp", "127.0.0.1:6379", "", []byte("test"))
	if err != nil {
		panic(err)
	}
	defer store.Close()

	server := &Server{service: service, sessionStore: store}

	r := gin.Default()
	r.LoadHTMLGlob("C:/code/weibo/html/*")

	r.Any("/login", server.login)
	r.Any("/register", server.register)
	r.GET("/weibo/publish", server.publishWeibo)
	r.GET("/weibo/delete", server.deleteWeibo)
	r.GET("/weibo/follow", server.follow)
	r.GET("/weibo/unfollow", server.unFollow)
	r.GET("/weibo/givelike", server.givelike)
	r.GET("/weibo/collect", server.collect)
	r.GET("/weibo/postComment", server.postComment)
	r.GET("/weibo/deleteComment", server.deleteComment)
	r.GET("/weibo/weiboList", server.weiboList)
	r.GET("/weibo/followersShow", server.followersShow)
	r.POST("/weibo/searchWeibo", server.searchWeibo)
	r.POST("/weibo/searchUser", server.searchUser)
	r.GET("/notification", server.notificationPage)
	r.GET("/")
	r.Static("/html", "C:/code/weibo/html")
	// r.POST("/weibo/weiboList", server.weiboList)

	r.Run()
}

type Server struct {
	service      *weibo.Service
	sessionStore sessions.Store
}

func (s *Server) login(c *gin.Context) {
	// GET请求时，显示表单内容
	if c.Request.Method == "GET" {
		c.HTML(200, "index.html", gin.H{"title": "登录"})
		return
	}

	// 非GET请求时，认为是POST请求，这时处理表单数据
	var account string
	err := func() error {
		account = c.PostForm("account")
		password := c.PostForm("password")
		if len(account) == 0 || len(password) == 0 {
			return errors.New("参数错误")
		}

		user, err := s.service.Login(account, password)
		if err != nil {
			return err
		}

		// user -> session
		s.saveUserToSession(c, user)
		return nil
	}()

	// 表单处理失败时，返回错误信息和表单内容
	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		// c.HTML(200, "index.html", gin.H{
		// 	"title":   "登录",
		// 	"account": account,
		// 	"err":     err,
		// })
		return
	}

	// 表单处理成功后，跳转到下一个页面
	c.Redirect(302, "/weibo/weiboList")
}

func (s *Server) register(c *gin.Context) {
	// user, err := s.service.Register(account, password)
	if c.Request.Method == "GET" {
		c.HTML(200, "register.html", gin.H{"title": "注册"})
		return
	}

	var account string
	err := func() error {
		account := c.PostForm("account")
		avatar := c.PostForm("avatar")
		password := c.PostForm("password")
		repassword := c.PostForm("repassword")
		if len(account) == 0 || len(avatar) == 0 || len(password) == 0 {
			return errors.New("参数错误")
		}

		if password != repassword {
			return errors.New("二次密码不正确")
		}

		user, err := s.service.Register(account, avatar, password)
		if err != nil {
			return err
		}

		s.saveUserToSession(c, user)
		return nil
	}()

	if err != nil {
		c.HTML(200, "register.html", gin.H{
			"title":   "注册",
			"account": account,
			"err":     err,
		})
		return
	}

	c.Redirect(302, "/weibo/weiboList")
}

func (s *Server) publishWeibo(c *gin.Context) {

	var err error
	defer func() {
		if err != nil {
			s.redirectToNotificationPageWithError(c, err)
		}
	}()

	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	content := c.Query("content")
	if len(content) == 0 {
		err = errors.New("微博不能为空")
		return
	}

	w := &weibo.Weibo{
		UserID:    user.ID,
		Account:   user.Account,
		Content:   content,
		CreatedAt: time.Now().Unix(),
	}

	err = s.service.PublishWeibo(user, w)
	if err != nil {
		return
	}

	c.HTML(200, "listNew.html", w)
}

func (s *Server) deleteWeibo(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			s.redirectToNotificationPageWithError(c, err)
		}
	}()

	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	weiboIDStr := c.Query("weiboID")
	weiboID, _ := strconv.ParseInt(weiboIDStr, 10, 64)
	if weiboID == 0 {
		err = errors.New("微博不存在")
		return
	}

	err = s.service.DeleteWeibo(user, weiboID)
	if err != nil {
		return
	}

	s.saveUserToSession(c, nil)
	return
}

func (s *Server) follow(c *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			s.redirectToNotificationPageWithError(c, err)
		}
	}()

	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	toUserIDStr := c.Query("id")
	toUserID, _ := strconv.ParseInt(toUserIDStr, 10, 64)
	if toUserID == 0 {
		err = errors.New("用户不存在")
		return
	}

	err = s.service.Follow(user, toUserID)
	if err != nil {
		return
	}

	s.saveUserToSession(c, nil)
	return
}

func (s *Server) unFollow(c *gin.Context) {
	var err error

	defer func() {
		if err != nil {
			s.redirectToNotificationPageWithError(c, err)
		}
	}()

	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	toUserIDStr := c.Query("id")
	toUserID, _ := strconv.ParseInt(toUserIDStr, 10, 64)
	if toUserID == 0 {
		err = errors.New("用户不存在")
		return
	}

	err = s.service.UnFollow(user, toUserID)
	if err != nil {
		return
	}

	s.saveUserToSession(c, nil)
	return
}

func (s *Server) givelike(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	err := func() error {
		weiboIDStr := c.Query("weiboID")
		weiboID, _ := strconv.ParseInt(weiboIDStr, 10, 64)
		if weiboID == 0 {
			return errors.New("微博不存在")
		}

		err := s.service.Givelike(user, weiboID)
		if err != nil {
			return err
		}

		s.saveUserToSession(c, nil)
		return nil
	}()

	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	c.Redirect(302, "/weibo/weiboList")
}

func (s *Server) collect(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	err := func() error {
		weiboIDStr := c.Query("weiboID")
		weiboID, _ := strconv.ParseInt(weiboIDStr, 10, 64)
		if weiboID == 0 {
			return errors.New("微博不存在")
		}

		err := s.service.Collect(user, weiboID)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	s.redirectToNotificationPageWithMessage(c, "收藏成功")
}

func (s *Server) postComment(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	err := func() error {
		weiboIDStr := c.Query("weiboID")
		weiboID, _ := strconv.ParseInt(weiboIDStr, 10, 64)
		if weiboID == 0 {
			return errors.New("微博不存在")
		}

		commentContent := c.Query("commentContent")
		err := s.service.PostComment(user, weiboID, commentContent)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	s.redirectToNotificationPageWithMessage(c, "评论成功")
}

func (s *Server) deleteComment(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	err := func() error {
		commentIDStr := c.Query("commentID")
		commentID, _ := strconv.ParseInt(commentIDStr, 10, 64)
		if commentID == 0 {
			return errors.New("微博不存在")
		}

		err := s.service.DeleteComment(user, commentID)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	s.redirectToNotificationPageWithMessage(c, "删除评论成功")
}

func (s *Server) weiboList(c *gin.Context) {

	user := s.getUserFromSession(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}

	var page int64 = 1
	pageStr := c.Query("page")
	if len(pageStr) > 0 {
		page, _ = strconv.ParseInt(pageStr, 10, 64)
		if page == 0 {
			page = 1
		}
	}

	user, followers, weibos, err := s.service.WeiboList(user, page, 15)
	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	c.HTML(200, "listNew.html", gin.H{
		"user":      user,
		"weibos":    weibos,
		"followers": followers,
	})
}

func (s *Server) followersShow(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}

	var page int64 = 1
	pageStr := c.Query("page")
	if len(pageStr) > 0 {
		page, _ = strconv.ParseInt(pageStr, 10, 64)
		if page == 0 {
			page = 1
		}
	}

	user, weibos, err := s.service.FollowersShow(user, page, 15)
	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	c.HTML(200, "home.html", gin.H{
		"user":   user,
		"weibos": weibos,
	})
}

func (s *Server) searchWeibo(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	var page int64 = 1
	pageStr := c.Query("page")
	if len(pageStr) > 0 {
		page, _ = strconv.ParseInt(pageStr, 10, 64)
		if page == 0 {
			page = 1
		}
	}

	accountOrContent := c.Query("accountOrContent")
	weibos, err := s.service.SearchWeibo(accountOrContent, page, 15)
	// weibos, err := s.service.SearchWeibo(accountOrContent, page, perPage)
	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	c.HTML(200, "listNew.html", gin.H{
		"user":   user,
		"weibos": weibos,
	})
}

func (s *Server) searchUser(c *gin.Context) {
	user := s.getUserFromSession(c)
	if user == nil {
		s.redirectToNotificationPageWithError(c, errors.New("先登录"))
		return
	}

	var page int64 = 1
	pageStr := c.Query("page")
	if len(pageStr) > 0 {
		page, _ = strconv.ParseInt(pageStr, 10, 64)
		if page == 0 {
			page = 1
		}
	}

	users, err := s.service.SearchUser("", page, 15)
	if err != nil {
		s.redirectToNotificationPageWithError(c, err)
		return
	}

	c.HTML(200, "list.html", users)
}

func (s *Server) notificationPage(c *gin.Context) {
	message, err := s.getNotificationFromSession(c)
	if err != nil {
		c.String(200, "发生错误: "+err.Error())
		return
	}

	if len(message) > 0 {
		c.String(200, "系统通知: "+message)
		return
	}

	c.Redirect(302, "/")
}

// err = service.Register(&weibo.User{
// 	Account:  "lhc1",
// 	Avatar:   "xxx.jpg",
// 	Password: "123123",
// })

// if err == nil {
// 	fmt.Println("用户注册成功")
// } else {
// 	fmt.Println(err)
// }
