package weibo

type Comment struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"user_id" db:"user_id"`
	WeiboID   int64  `json:"weibo_id" db:"weibo_id"`
	Content   string `json:"content" db:"content"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}
