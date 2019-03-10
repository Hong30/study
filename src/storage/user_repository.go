package storage

import (
	"database/sql"
	"weibo"

	"github.com/jmoiron/sqlx"
)

// 用户仓库
type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
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
