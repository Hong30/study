package weibo

type Givelike struct {
	UserID    int64 `json:"user_id" db:"user_id"`
	WeiboID   int64 `json:"weibo_id" db:"weibo_id"`
	CreatedAt int64 `json:"created_at" db:"created_at"`
}
