package weibo

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"time"

	"github.com/pkg/errors"
)

// 微博服务
type Service struct {
	userRepo     UserRepository
	timelineRepo TimeLineRepository
	weiboRepo    WeiboRepository
}

func NewService(userRepo UserRepository, weiboRepo WeiboRepository, timelineRepo TimeLineRepository) *Service {
	return &Service{
		userRepo:     userRepo,
		timelineRepo: timelineRepo,
		weiboRepo:    weiboRepo,
	}
}

// 注册功能
func (s *Service) Register(account, avatar, password string) (*User, error) {
	existsUser, err := s.userRepo.GetUserByAccount(account)
	if err != nil {
		return nil, errors.Wrap(err, "查询同名账号失败")
	}

	if existsUser != nil {
		return nil, errors.New("账号名已经被使用了")
	}

	user := &User{
		Account:   account,
		Avatar:    avatar,
		CreatedAt: time.Now().Unix(),
	}

	// 生成随机的盐
	buf := make([]byte, 4)
	rand.Read(buf)
	user.Salt = hex.EncodeToString(buf)

	// 用密码加盐生成哈希
	hash := sha1.New()
	hash.Write([]byte(password + user.Salt))
	buf = hash.Sum(nil)
	user.Password = hex.EncodeToString(buf)

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, errors.Wrap(err, "保存用户信息失败")
	}

	return user, nil
}

func (s *Service) Login(account, password string) (user *User, err error) {
	existsUser, err := s.userRepo.GetUserByAccount(account)
	if err != nil {
		return nil, err
	}

	if existsUser == nil {
		return nil, errors.New("用户不存在")
	}

	hash := sha1.New()
	hash.Write([]byte(password + existsUser.Salt))
	if existsUser.Password == hex.EncodeToString(hash.Sum(nil)) {
		return existsUser, nil
	}

	return nil, errors.New("密码错误")
}

func (s *Service) Follow(user *User, targetUserID int64) error {
	targetUser, err := s.userRepo.GetUserByID(targetUserID)
	if err != nil {
		return errors.Wrap(err, "查询目标用户失败")
	}
	if targetUser == nil {
		return errors.New("要关注的用户不存在")
	}

	following, err := s.userRepo.GetFollowing(user.ID, targetUserID)
	if err != nil {
		return errors.Wrap(err, "无法查询当前用户是否已关注过目标用户")
	}
	if following != nil {
		return errors.New("关注过目标用户")
	}

	following = &Following{
		FromUserID: user.ID,
		ToUserID:   targetUserID,
	}
	if err := s.userRepo.CreateFollowing(following); err != nil {
		return errors.Wrap(err, "关注信息保存失败")
	}

	if err := s.userRepo.AddFollowingNumByUserID(user.ID, 1); err != nil {
		return errors.Wrap(err, "用户所关注的人数增加失败")
	}

	timeLines, err := s.timelineRepo.GetRecentlyWeiboIDsByUserID(targetUserID, 30)
	if err != nil {
		return errors.Wrap(err, "目标用户最近的微博获取失败")
	}

	for _, timeline := range timeLines {
		timeline.UserID = user.ID
	}
	if err := s.timelineRepo.BatchCreateTimeLines(timeLines); err != nil {
		return errors.Wrap(err, "当前用户的时间线更新失败")
	}

	if err := s.userRepo.AddFollowerNumByUserID(targetUserID, 1); err != nil {
		return errors.Wrap(err, "用户的粉丝数增加失败")
	}

	return nil
}

func (s *Service) UnFollow(user *User, toUserID int64) error {
	following, err := s.userRepo.GetFollowing(user.ID, toUserID)
	if err != nil {
		return errors.Wrap(err, "无法查询当前用户是否已关注过目标用户")
	}
	if following != nil {
		return errors.New("关注过目标用户")
	}

	// 删除关注关系
	if err := s.userRepo.DeleteFollowing(following); err != nil {
		return errors.Wrap(err, "删除关注关系失败")
	}

	// 更新关注数和粉丝数
	if err := s.userRepo.AddFollowingNumByUserID(user.ID, -1); err != nil {
		return errors.Wrap(err, "用户关注人数减少失败")
	}

	if err := s.userRepo.AddFollowingNumByUserID(toUserID, -1); err != nil {
		return errors.Wrap(err, "用户的粉丝取消关注失败")
	}

	// 从当前用户的timeline中删除被关注者的微博
	if err := s.timelineRepo.DeleteWeiboByUserIDAndWeiboUserID(user.ID, toUserID); err != nil {
		return errors.Wrap(err, "从当前用户的timeline中删除被关注者的微博失败")
	}

	return nil
}

func (s *Service) PublishWeibo(user *User, weibo *Weibo) error {
	// 数据有效性的检查
	//weiboID, err := s.userRepo.ExamineData(user.ID)

	// 把微博插入到数据库， 成功后获取微博的id
	weiboID, err := s.weiboRepo.InsertWeibo(weibo)
	if err != nil {
		return errors.Wrap(err, "获取微博id失败")
	}

	// 在自己的timeline中增加这条微博
	newTimeline := &TimeLine{
		UserID:         user.ID,
		WeiboUserID:    user.ID,
		WeiboID:        weiboID,
		WeiboCreatedAt: weibo.CreatedAt,
	}
	if err := s.timelineRepo.CreateTimeLine(newTimeline); err != nil {
		return errors.Wrap(err, "保存微博到当前用户的timeline失败")
	}

	// 增加自己的微博数量
	if err := s.userRepo.AddWeiboNumByUserID(user.ID, 1); err != nil {
		return errors.Wrap(err, "用户的微博数增加失败")
	}

	// 查询当前用户的粉丝， 这里只查id就可以了
	followings, err := s.userRepo.GetUserFollowers(user.ID)

	// 遍历粉丝，向他们的timeline中增加这条微博
	for _, following := range followings {
		newTimeline.UserID = following.FromUserID
		if err := s.timelineRepo.CreateTimeLine(newTimeline); err != nil {
			log.Printf("时间线插入失败: %+v\n", newTimeline)
		}
	}

	return nil
}

func (s *Service) DeleteWeibo(user *User, weiboID int64) error {
	//从数据库中查找是否有这条微博id
	weibo, err := s.weiboRepo.GetWeiboByID(weiboID)
	if err != nil {
		return errors.Wrap(err, "获取微博id失败")
	}
	if weibo == nil || weibo.UserID != user.ID {
		return errors.New("微博不存在")
	}

	//在自己的微博列表中删除这条微博
	if err = s.weiboRepo.DeleteWeibo(weibo); err != nil {
		return errors.Wrap(err, "删除微博失败")
	}

	//减少发布的微博数量
	if err = s.userRepo.AddWeiboNumByUserID(user.ID, -1); err != nil {
		return errors.Wrap(err, "用户的微博数减少失败")
	}

	//把全部粉丝中的这条微博删掉
	followers, err := s.userRepo.GetUserFollowers(user.ID)
	if err != nil {
		return errors.Wrap(err, "获取粉丝信息失败")
	}

	for _, follower := range followers {
		if err := s.timelineRepo.DeleteWeiboByUserIDAndWeiboID(follower.FromUserID, weiboID); err != nil {
			log.Printf("删除用户 %d 的timeline中的微博 %d 失败\n", follower.FromUserID, weiboID)
		}
	}

	if err := s.timelineRepo.DeleteWeiboByUserIDAndWeiboID(user.ID, weiboID); err != nil {
		log.Printf("删除用户 %d 的timeline中的微博 %d 失败\n", user.ID, weiboID)
	}
	return nil
}

func (s *Service) Givelike(user *User, weiboID int64) error {
	weibo, err := s.weiboRepo.GetWeiboByID(weiboID)
	if err != nil {
		return errors.Wrap(err, "查询微博失败")
	}

	if weibo == nil {
		return errors.Wrap(err, "这条微博不存在")
	}

	givelike, err := s.weiboRepo.GetGivelikeByUseIDAndWeiboID(user.ID, weibo.ID)
	if err != nil {
		return errors.Wrap(err, "点赞错误")
	}

	if givelike != nil {
		return errors.New("已点赞过该微博")
	}

	if err := s.weiboRepo.AddLikeNumByWeiboID(weibo.ID, 1); err != nil {
		return errors.Wrap(err, "点赞失败")
	}

	newGivelike := &Givelike{
		UserID:  user.ID,
		WeiboID: weibo.ID,
	}

	if err := s.weiboRepo.CreateGivelike(newGivelike); err != nil {
		return errors.Wrap(err, "保存点赞记录到当前用户微博失败")
	}

	return nil
}

func (s *Service) Collect(user *User, weiboID int64) error {
	// 判断微博是否存在
	weibo, err := s.weiboRepo.GetWeiboByID(weiboID)
	if err != nil {
		return errors.Wrap(err, "查询微博失败")
	}

	if weibo == nil {
		return errors.Wrap(err, "这条微博不存在")
	}

	// 是否有收藏记录
	collect, err := s.weiboRepo.CollectByUseIDAndWeiboID(user.ID, weibo.ID)
	if err != nil {
		return errors.Wrap(err, "收藏失败")
	}

	if collect != nil {
		return errors.New("已收藏过该微博")
	}

	newCollect := &Collect{
		UserID:  user.ID,
		WeiboID: weibo.ID,
	}

	// 保存收藏记录
	if err := s.weiboRepo.CreateCollect(newCollect); err != nil {
		return errors.Wrap(err, "保存收藏记录到当前用户微博失败")
	}

	return nil
}

// 收藏： n个用户-n个微博(many to many) 需要关联表
// 评论： 1条微博-n个评论(one to many)

func (s *Service) PostComment(user *User, weiboID int64, commentContent string) error {
	// 判断微博存在与否
	weibo, err := s.weiboRepo.GetWeiboByID(weiboID)
	if err != nil {
		return errors.Wrap(err, "查询微博失败")
	}

	if weibo == nil {
		return errors.Wrap(err, "这条微博不存在")
	}

	// 在微博中增加评论记录
	if err := s.weiboRepo.AddCommentNumByWeiboID(weiboID, 1); err != nil {
		return errors.Wrap(err, "评论失败")
	}

	newComment := &Comment{
		UserID:    user.ID,
		WeiboID:   weibo.ID,
		Content:   commentContent,
		CreatedAt: time.Now().Unix(),
	}
	// 保存评论记录到库
	if err := s.weiboRepo.CreateComment(newComment); err != nil {
		return errors.Wrap(err, "保存评论记录到当前用户微博失败")
	}

	return nil
}

func (s *Service) DeleteComment(user *User, commentID int64) error {
	comment, err := s.weiboRepo.GetCommentByID(commentID)
	if err != nil {
		return errors.Wrap(err, "获取微博id失败")
	}
	if comment == nil {
		return errors.New("微博不存在")
	}

	// 当前用户是否可以删除这条评论
	if user.ID != comment.UserID {
		return errors.New("不是你的微博不可以删除")
	}

	if err = s.weiboRepo.DeleteComment(commentID); err != nil {
		return errors.Wrap(err, "删除评论失败")
	}

	if err = s.weiboRepo.AddCommentNumByWeiboID(comment.WeiboID, -1); err != nil {
		return errors.Wrap(err, "微博评论数减少失败")
	}

	return nil
}

func (s *Service) WeiboList(user *User, page, perPage int64) (*User, []*Follower, []*WeiboWithUser, error) {
	// INNER JOIN
	// LEFT JOIN
	// RIGHT JOIN
	// SELECT w.* FROM `weibos` w INNER JOIN `timelines` t ON w.id = t.weibo_id AND t.user_id = ? ORDER BY t.created_at DESC LIMIT ?,?

	offset := (page - 1) * perPage
	weibos, err := s.weiboRepo.GetWeibosByUserTimelines(user.ID, offset, perPage)
	log.Println("weibos:", weibos)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "查询微博失败")
	}

	if len(weibos) == 0 {
		return user, nil, []*WeiboWithUser{}, nil
	}

	user, err = s.userRepo.GetUserByID(user.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	followers, err := s.userRepo.GetUserFollowers2(user.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	return user, followers, weibos, nil
}

func (s *Service) FollowersShow(user *User, page, perPage int64) (*User, []*WeiboWithUser, error) {
	offset := (page - 1) * perPage
	weibos, err := s.weiboRepo.GetWeibosByUserTimelines(user.ID, offset, perPage)
	log.Println("weibos:", weibos)
	if err != nil {
		return nil, nil, errors.Wrap(err, "查询微博失败")
	}

	if len(weibos) == 0 {
		return user, []*WeiboWithUser{}, nil
	}

	user, err = s.userRepo.GetUserByID(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, weibos, nil
}

//搜索微博
func (s *Service) SearchWeibo(accountOrContent string, page, perPage int64) ([]*Weibo, error) {
	offset := (page - 1) * perPage
	weibos, err := s.weiboRepo.GetWeibosByAccountOrContent(accountOrContent, offset, perPage)
	if err != nil {
		return nil, errors.Wrap(err, "搜索微博失败")
	}

	if len(weibos) != 0 {
		return []*Weibo{}, nil
	}

	return weibos, nil
	//s.weiboRepo.CountWeibos(accountOrContent)
}

func (s *Service) SearchUser(account string, page, perPage int64) ([]*User, error) {
	offset := (page - 1) * perPage
	users, err := s.userRepo.GetUsersByAccount(account, offset, perPage)
	if err != nil {
		return nil, errors.Wrap(err, "搜索用户失败")
	}
	return users, nil
}

// func (s *Service) GetUserProfile(userID int64) (*User, error) {
// 	user, err := s.userRepo.GetUserByID(userID)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "查询目标用户失败")
// 	}

// 	if user == nil {
// 		return nil, errors.New("用户不存在")
// 	}

// 	return user, nil
// }

// func (s *Service) Follow(user *User, targetUserID int64) error {
// 	targetUser, err := s.userRepo.GetUserByID(targetUserID)
// 	if err != nil {
// 		return errors.Wrap(err, "查询目标用户失败")
// 	}
// 	// 判断目标是否存在
// 	if targetUser == nil {
// 		return errors.New("要关注的用户不存在")
// 	}

// 	// 判断是否已经关注过目标用户
// 	following, err := s.userRepo.GetFollowing(user.ID, targetUserID)
// 	if err != nil {
// 		return errors.Wrap(err, "无法查询当前用户是否已关注过目标用户")
// 	}
// 	if following != nil {
// 		return errors.New("已经关注过目标用户")
// 	}

// 	// 事务 transaction

// 	// 创建新的关注信息，并记录到数据库
// 	following = &Following{
// 		FromUserID: user.ID,
// 		ToUserID:   targetUserID,
// 	}
// 	if err := s.userRepo.CreateFollowing(following); err != nil {
// 		return errors.Wrap(err, "保存关注信息失败")
// 	}

// 	if err := s.userRepo.AddFollowingNumByUserID(user.ID, 1); err != nil {
// 		return errors.Wrap(err, "增加用户所关注的人数失败")
// 	}

// 	timeLines, err := s.timelineRepo.GetRecentlyWeiboIDsByUserID(targetUserID)
// 	if err != nil {
// 		return errors.Wrap(err, "获取目标用户最近的微博失败")
// 	}

// 	for _, timeline := range timeLines {
// 		timeline.UserID = user.ID
// 	}
// 	if err := s.timelineRepo.BatchCreateTimeLines(timeLines); err != nil {
// 		return errors.Wrap(err, "更新当前用户的时间线失败")
// 	}

// 	if err := s.userRepo.AddFollowerNumByUserID(targetUserID, 1); err != nil {
// 		return errors.Wrap(err, "增加用户的粉丝数失败")
// 	}

// 	return nil
// }

// 	log.Panicln("登陆成功")
// 	return
// }

// func (s *Service) Login(account, password string) (user *User, err error) {
// 	existsUser, err := s.userRepo.GetUserByAccount(account)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if existsUser == nil {
// 		return nil, errors.New("用户不存在")
// 	}

// 	hash := sha1.New()
// 	hash.Write([]byte(password + user.Salt))
// 	if user.Password == hex.EncodeToString(hash.Sum(nil)) {
// 		return user, nil
// 	}

// 	return nil, errors.New("密码不正确")
// }
