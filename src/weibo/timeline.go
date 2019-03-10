package weibo

type TimeLine struct {
	ID int64 `json:"id" db:"id"`
	// timeline的拥有者的id
	UserID int64 `json:"user_id" db:"user_id"`
	// 微博发布者的id
	WeiboUserID int64 `json:"weibo_user_id" db:"weibo_user_id"`
	// weibo的id
	WeiboID int64 `json:"weibo_id" db:"weibo_id"`
	// weibo的发布时间
	WeiboCreatedAt int64 `json:"weibo_created_at" db:"weibo_created_at"`
}
