package weibo

type Weibo struct {
	ID         int64  `json:"id" db:"id"`
	UserID     int64  `json:"user_id" db:"user_id"`
	Account    string `json:"account" db:"account"` // 冗余字段
	Content    string `json:"content" db:"content"`
	LikeNum    int32  `json:"like_num" db:"like_num"`
	CommentNum int32  `json:"comment_num" db:"comment_num"` // 冗余字段
	CreatedAt  int64  `json:"created_at" db:"created_at"`
}

type WeiboWithUser struct {
	Weibo
	Avatar string `json:"avatar" db:"avatar"`
}
