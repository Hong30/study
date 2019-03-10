package weibo

import "testing"

type MockUserRepository struct{}

func (r *MockUserRepository) GetUserByAccount(account string) (*User, error) {
	if account == "exists" {
		return &User{ID: 999, Account: "exists"}, nil
	}
	return nil, nil
}
func (r *MockUserRepository) CreateUser(user *User) error {
	user.ID = 1
	return nil
}
func (r *MockUserRepository) GetUserByID(userID int64) (*User, error)               { return nil, nil }
func (r *MockUserRepository) AddFollowingNumByUserID(userID int64, num int32) error { return nil }
func (r *MockUserRepository) AddFollowerNumByUserID(userID int64, num int32) error  { return nil }
func (r *MockUserRepository) GetFollowing(fromUserID, toUserID int64) (*Following, error) {
	return nil, nil
}
func (r *MockUserRepository) CreateFollowing(following *Following) error          { return nil }
func (r *MockUserRepository) DeleteFollowing(following *Following) error          { return nil }
func (r *MockUserRepository) AddWeiboNumByUserID(userID int64, num int32) error   { return nil }
func (r *MockUserRepository) GetUserFollowers(userID int64) ([]*Following, error) { return nil, nil }
func (r *MockUserRepository) GetUsersByAccount(account string, offset, perPage int64) ([]*User, error) {
	return nil, nil
}

func TestRegister(t *testing.T) {
	service := NewService(&MockUserRepository{}, nil, nil)
	_, err := service.Register("exists", "xxx.jpg", "123123")
	if err == nil {
		t.Fatal("账号是否重复的判断有问题")
	}

	user, err := service.Register("hc", "xxx.jpg", "123123")
	if err != nil {
		t.Fatal("注册失败", err)
	}
	if user.ID == 0 {
		t.Fatal("注册失败,id问题", user)
	}
	t.Log("注册成功", user)
}

func TestLogin(t *testing.T) {
	Service := NewService(&MockUserRepository{}, nil, nil)
	_, err := Service.Login("exists", "...")
	if err == nil {
		t.Fatal("账号不为空")
	}

	user, err := Service.Login(",,,", "...")
	if err != nil {
		t.Fatal("登录失败", err)
	}
	if user.ID == 0 {
		t.Fatal("登录失败", user)
	}
	t.Log("登录成功", user)
}

// func TestFollow(t *testing.T) {
// 	Service := NewService(&MockUserRepository{}, nil, nil)
// 	_, err := Service.Follow("exists", "...")
// 	if err == nil {
// 		t.Fatal("账号不为空")
// 	}

// 	following, err := Service.Follow(user, targetUserID)
// 	if err != nil {
// 		t.Fatal("关注错误", err)
// 	}
// 	if following == 0 {
// 		t.Fatal("关注失败", user)
// 	}

// 	t.Log("登录成功", nil )
// }
