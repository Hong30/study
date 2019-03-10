package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"weibo"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

var _ weibo.UserRepository = new(UserRepository)

// 用户仓库
type UserRepository struct {
	db          *sqlx.DB
	redisClient redis.Conn
}

// todo: 增加redis client
func NewUserRepository(db *sqlx.DB, redisClient redis.Conn) *UserRepository {
	return &UserRepository{db: db, redisClient: redisClient}
}

// 通过账号名查找用户
func (ur *UserRepository) GetUserByAccount(account string) (*weibo.User, error) {
	var user weibo.User
	if err := ur.db.Get(&user, "SELECT * FROM `users` WHERE `account` = ?", account); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// 创建用户
func (ur *UserRepository) CreateUser(user *weibo.User) error {
	result, err := ur.db.NamedExec("INSERT INTO `users`(account, avatar, password, salt, created_at) VALUES(:account, :avatar, :password, :salt, :created_at)", user)
	if err != nil {
		return err
	}

	user.ID, err = result.LastInsertId()
	return err
}

// 根据用户id查询相应用户信息
func (ur *UserRepository) GetUserByID(userID int64) (*weibo.User, error) {
	var user weibo.User
	if err := ur.db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// 增加用户所关注的人数
func (ur *UserRepository) AddFollowingNumByUserID(userID int64, num int32) error {
	_, err := ur.db.Exec("UPDATE `users` SET following_num = following_num + ? WHERE id = ?", num, userID)
	return err
}

// 增加用户的粉丝数
func (ur *UserRepository) AddFollowerNumByUserID(userID int64, num int32) error {
	_, err := ur.db.Exec("UPDATE `users` SET follower_num = follower_num + ? WHERE id = ?", num, userID)
	return err
}

// 查询某个用户关注另一个用户的记录
func (ur *UserRepository) GetFollowing(fromUserID, toUserID int64) (*weibo.Following, error) {
	var following weibo.Following
	if err := ur.db.Get(&following, "SELECT * FROM `following` WHERE `from_user_id` = ? AND `to_user_id` = ?", fromUserID, toUserID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &following, nil
}

// 记录关注信息
func (ur *UserRepository) CreateFollowing(following *weibo.Following) error {
	_, err := ur.db.NamedExec("INSERT INTO `following`(from_user_id, to_user_id, created_at) VALUES(:from_user_id, :to_user_id, :created_at)", following)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("followers:%d", following.ToUserID)
	_, err = ur.redisClient.Do("DEL", key)
	if err != nil {
		return err
	}

	return err
}

// 删除关注信息
func (ur *UserRepository) DeleteFollowing(following *weibo.Following) error {
	_, err := ur.db.Exec("DELETE FROM `following` WHERE from_user_id = ? AND to_user_id = ?", following.FromUserID, following.ToUserID)

	key := fmt.Sprintf("followers:%d", following.ToUserID)
	_, err = ur.redisClient.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

// 增加用户所发布的微博数量
func (ur *UserRepository) AddWeiboNumByUserID(userID int64, num int32) error {
	_, err := ur.db.Exec("UPDATE `users` SET weibo_num = weibo_num + ? WHERE id = ?", num, userID)
	return err
}

// 获取用户所粉丝
func (ur *UserRepository) GetUserFollowers(userID int64) ([]*weibo.Following, error) {
	followings := []*weibo.Following{}
	if err := ur.db.Select(&followings, "SELECT * FROM `following` WHERE to_user_id = ?", userID); err != nil {
		return nil, err
	}
	return followings, nil
}

func (ur *UserRepository) GetUsersByAccount(account string, offset, perPage int64) ([]*weibo.User, error) {
	users := []*weibo.User{}
	if err := ur.db.Select(&users, "SELECT * FROM `users` WHERE `account` = ?", account); err != nil {
		return nil, err
	}
	return users, nil
}

func (ur *UserRepository) GetUserFollowers2(userID int64) ([]*weibo.Follower, error) {
	followers := []*weibo.Follower{}
	key := fmt.Sprintf("follower:%d", userID)

	// 使用redis client, 先从redis中读数据， 如果没有则从mysql读并写入缓存
	data, err := redis.Bytes(ur.redisClient.Do("GET", key))
	if err == nil {
		if err = json.Unmarshal(data, &followers); err != nil {
			return nil, err
		}
		return followers, nil
	}

	if err != redis.ErrNil {
		return nil, err
	}

	query := `
		SELECT u.id, u.account, u.avatar FROM users u
		INNER JOIN following f ON f.from_user_id = u.id
		WHERE f.to_user_id = ?
	`
	if err := ur.db.Select(&followers, query, userID); err != nil {
		return nil, err
	}

	data, err = json.Marshal(followers)
	if err != nil {
		return nil, err
	}

	_, err = ur.redisClient.Do("SETEX", key, 300, data)
	if err != nil {
		return nil, err
	}

	return followers, nil
}
