package weibo

// 用户
type User struct {
	ID           int64  `json:"id" db:"id"`
	Account      string `json:"account" db:"account"`
	Avatar       string `json:"avatar" db:"avatar"`
	Password     string `json:"-" db:"password"`
	Salt         string `json:"-" db:"salt"`
	FollowingNum int32  `json:"following_num" db:"following_num"`
	FollowerNum  int32  `json:"follower_num" db:"follower_num"`
	WeiboNum     int32  `json:"weibo_num" db:"weibo_num"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
}

type Follower struct {
	ID      int64  `json:"id" db:"id"`
	Account string `json:"account" db:"account"`
	Avatar  string `json:"avatar" db:"avatar"`
}
