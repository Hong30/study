package weibo

type UserRepository interface {
	GetUserByAccount(account string) (*User, error)
	CreateUser(user *User) error
}
