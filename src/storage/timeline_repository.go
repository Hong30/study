package storage

import (
	"weibo"

	"github.com/jmoiron/sqlx"
)

type TimeLineRepository struct {
	db *sqlx.DB
}

func NewTimeLineRepository(db *sqlx.DB) *TimeLineRepository {
	return &TimeLineRepository{db: db}
}

// 查询某个用户最近七天的微博id
func (tl *TimeLineRepository) GetRecentlyWeiboIDsByUserID(userID int64, limit int32) ([]*weibo.TimeLine, error) {
	timeLines := []*weibo.TimeLine{}
	if err := tl.db.Select(&timeLines, "SELECT * FROM `timeline` WHERE `user_id` = ? ORDER BY created_id DESC LIMIT ?", userID, limit); err != nil {
		return nil, err
	}
	return timeLines, nil
}

// 批量把一批微博id插入到某个用户的timeline中
func (tl *TimeLineRepository) BatchCreateTimeLines(timelines []*weibo.TimeLine) error {
	for _, timeline := range timelines {
		_, err := tl.db.NamedExec("INSERT INTO `timeline`(id, user_id, weibo_user_id, weibo_id, weibo_created_at) VALUES(:id, :user_id, :weibo_user_id, :weibo_id, :weibo_created_at)", timeline)
		if err != nil {
			return err
		}
	}
	return nil
}

// 删除某个用户的timeline中某人发的微博
func (tl *TimeLineRepository) DeleteWeiboByUserIDAndWeiboUserID(userID int64, weiboUserID int64) error {
	_, err := tl.db.Exec("DELETE FROM `timeline` WHERE user_id = ? AND weibo_user_id = ?", userID, weiboUserID)
	return err
}

// 删除某个用户的timeline中的某条微博
func (tl *TimeLineRepository) DeleteWeiboByUserIDAndWeiboID(userID int64, weiboID int64) error {
	_, err := tl.db.Exec("DELETE FROM `timeline` WHERE user_id = ? AND weibo_id = ?", userID, weiboID)
	return err
}

// 插入新的timeline
func (tl *TimeLineRepository) CreateTimeLine(timeline *weibo.TimeLine) error {
	_, err := tl.db.NamedExec("INSERT INTO `timeline`(user_id, weibo_user_id, weibo_id, weibo_created_at) VALUES(:user_id, :weibo_user_id, :weibo_id, :weibo_created_at)", timeline)
	if err != nil {
		return err
	}
	return nil
}
