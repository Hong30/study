package weibo

type UserRepository interface {
	GetUserByAccount(account string) (*User, error)
	CreateUser(user *User) error
	// 根据用户id查询相应用户信息
	GetUserByID(userID int64) (*User, error)
	// 增加用户所关注的人数
	AddFollowingNumByUserID(userID int64, num int32) error
	// 增加用户的粉丝数
	AddFollowerNumByUserID(userID int64, num int32) error

	// 查询某个用户关注另一个用户的记录
	GetFollowing(fromUserID, toUserID int64) (*Following, error)
	// 记录关注信息
	CreateFollowing(following *Following) error
	// 删除关注信息
	DeleteFollowing(following *Following) error
	// 增加微博数量
	AddWeiboNumByUserID(userID int64, num int32) error

	// 获取用户所粉丝
	GetUserFollowers(userID int64) ([]*Following, error)

	GetUserFollowers2(userID int64) ([]*Follower, error)

	GetUsersByAccount(account string, offset, perPage int64) ([]*User, error)
}

type WeiboRepository interface {
	GetWeiboByID(weiboID int64) (*Weibo, error)
	InsertWeibo(weibo *Weibo) (int64, error)
	DeleteWeibo(weibo *Weibo) error
	CreateGivelike(giveLike *Givelike) error
	AddLikeNumByWeiboID(weiboID int64, num int32) error
	GetGivelikeByUseIDAndWeiboID(userID int64, weiboID int64) (*Givelike, error)
	CollectByUseIDAndWeiboID(userID int64, weiboID int64) (*Collect, error)
	CreateCollect(collect *Collect) error
	CreateComment(comment *Comment) error
	AddCommentNumByWeiboID(weiboID int64, num int32) error
	GetCommentByID(commentID int64) (*Comment, error)
	DeleteComment(weiboID int64) error
	// 根据用户Timelines找微博
	GetWeibosByUserTimelines(userID int64, offset, limit int64) ([]*WeiboWithUser, error)
	// 根据账号或者内容搜索微博
	GetWeibosByAccountOrContent(accountOrContent string, offset, perPage int64) ([]*Weibo, error)
}

type TimeLineRepository interface {
	// 查询某个用户最近七天的微博id
	GetRecentlyWeiboIDsByUserID(userID int64, limit int32) ([]*TimeLine, error)
	// 批量把一批微博id插入到某个用户的timeline中
	BatchCreateTimeLines(timelines []*TimeLine) error
	// 删除某个用户的timeline中某人发的微博
	DeleteWeiboByUserIDAndWeiboUserID(userID int64, weiboUserID int64) error
	// 删除某个用户的timeline中的某条微博
	DeleteWeiboByUserIDAndWeiboID(userID int64, weiboID int64) error
	// 插入新的timeline
	CreateTimeLine(timeline *TimeLine) error
}
