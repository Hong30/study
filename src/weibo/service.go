package weibo

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/pkg/errors"
)

// 微博服务
type Service struct {
	userRepo UserRepository
}

func NewService(userRepo UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// 注册功能
func (s *Service) Register(user *User) error {
	existsUser, err := s.userRepo.GetUserByAccount(user.Account)
	if err != nil {
		return errors.Wrap(err, "查询同名账号失败")
	}

	if existsUser != nil {
		return errors.New("账号名已经被使用了")
	}

	// 生成随机的盐
	buf := make([]byte, 4)
	rand.Read(buf)
	user.Salt = hex.EncodeToString(buf)

	// 用密码加盐生成哈希
	hash := sha1.New()
	hash.Write([]byte(user.Password + user.Salt))
	buf = hash.Sum(nil)
	user.Password = hex.EncodeToString(buf)

	user.CreatedAt = time.Now().Unix()
	if err := s.userRepo.CreateUser(user); err != nil {
		return errors.Wrap(err, "保存用户信息失败")
	}

	return nil
}
