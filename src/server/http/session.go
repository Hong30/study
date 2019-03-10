package main

import (
	"errors"
	"log"
	"weibo"

	"github.com/gin-gonic/gin"
)

func (s *Server) getUserFromSession(c *gin.Context) *weibo.User {
	session, _ := s.sessionStore.Get(c.Request, "weibo")
	user, ok := session.Values["user"]
	if !ok {
		return nil
	}
	return user.(*weibo.User)
}

func (s *Server) saveUserToSession(c *gin.Context, user *weibo.User) {
	session, _ := s.sessionStore.Get(c.Request, "weibo")
	session.Values["user"] = user
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println(err)
	}
}

func (s *Server) redirectToNotificationPageWithError(c *gin.Context, err error) {
	session, _ := s.sessionStore.Get(c.Request, "weibo")
	session.Values["err"] = err.Error()
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println(err)
	}
	c.Redirect(302, "/notification")
}

func (s *Server) redirectToNotificationPageWithMessage(c *gin.Context, message string) {
	session, _ := s.sessionStore.Get(c.Request, "weibo")
	session.Values["message"] = message
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Println(err)
	}
	c.Redirect(302, "/notification")
}

func (s *Server) getNotificationFromSession(c *gin.Context) (message string, err error) {
	session, _ := s.sessionStore.Get(c.Request, "weibo")
	defer session.Save(c.Request, c.Writer)

	e := session.Values["err"]
	if e != nil {
		err = errors.New(e.(string))
		delete(session.Values, "err")
		return
	}

	m := session.Values["message"]
	if m != nil {
		message = m.(string)
		delete(session.Values, "message")
	}
	return
}
