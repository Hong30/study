package weibo

type Following struct {
	// 关注者的id
	FromUserID int64 `json:"from_user_id" db:"from_user_id"`
	// 被关注者的id
	ToUserID int64 `json:"to_user_id" db:"to_user_id"`
	// 关注时间
	CreatedAt int64 `json:"created_at" db:"created_at"`
}
