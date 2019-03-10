package weibo

// 用户
type User struct {
	ID        int64  `json:"id" db:"id"`
	Account   string `json:"account" db:"account"`
	Avatar    string `json:"avatar" db:"avatar"`
	Password  string `json:"password" db:"password"`
	Salt      string `json:"salt" db:"salt"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}
