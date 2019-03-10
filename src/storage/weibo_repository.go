package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"weibo"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

var _ weibo.WeiboRepository = new(WeiboRepository)

// 仓库
type WeiboRepository struct {
	db          *sqlx.DB
	redisClient redis.Conn
}

func NewWeiboRepository(db *sqlx.DB, redisClient redis.Conn) *WeiboRepository {
	return &WeiboRepository{db: db, redisClient: redisClient}
}

//根据id查找微博
func (wb *WeiboRepository) GetWeiboByID(weiboID int64) (*weibo.Weibo, error) {
	var weibo weibo.Weibo
	if err := wb.db.Get(&weibo, "SELECT * FROM `weibos` WHERE `id` = ?", weiboID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &weibo, nil
}

//
func (wb *WeiboRepository) InsertWeibo(weibo *weibo.Weibo) (int64, error) {
	_, err := wb.db.NamedExec("INSERT INTO `weibos`(id, user_id) VALUES(:id, :user_id)", weibo)
	return weibo.ID, err
}

// 删除微博
func (wb *WeiboRepository) DeleteWeibo(weibo *weibo.Weibo) error {
	_, err := wb.db.Exec("DELETE FROM `weibos` WHERE id = ?", weibo.ID)
	return err
}

// 保存点赞记录
func (wb *WeiboRepository) CreateGivelike(givelike *weibo.Givelike) error {
	_, err := wb.db.NamedExec("INSERT INTO `givelike`(user_id, weibo_id) VALUES(:user_id, :weibo_id)", givelike)
	return err
}

// 增加点赞数
func (wb *WeiboRepository) AddLikeNumByWeiboID(weiboID int64, num int32) error {
	_, err := wb.db.Exec("UPDATE `weibos` SET like_num = like_num + ? WHERE id = ?", num, weiboID)
	return err
}

func (wb *WeiboRepository) GetGivelikeByUseIDAndWeiboID(userID int64, weiboID int64) (*weibo.Givelike, error) {
	var givelike weibo.Givelike
	if err := wb.db.Get(&givelike, "SELECT * FROM `givelike` WHERE `user_id` = ? AND weibo_id = ?", userID, weiboID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &givelike, nil
}

func (wb *WeiboRepository) CollectByUseIDAndWeiboID(userID int64, weiboID int64) (*weibo.Collect, error) {
	var collect weibo.Collect
	if err := wb.db.Get(&collect, "SELECT * FROM `collect` WHERE `user_id` = ? AND weibo_id = ?", userID, weiboID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &collect, nil
}

func (wb *WeiboRepository) CreateCollect(collect *weibo.Collect) error {
	_, err := wb.db.NamedExec("INSERT INTO `collect`(user_id, weibo_id, created_at) VALUES(:user_id, :weibo_id, :created_at)", collect)
	if err != nil {
		return err
	}
	return nil
}

func (wb *WeiboRepository) CreateComment(comment *weibo.Comment) error {
	_, err := wb.db.NamedExec("INSERT INTO `weibos`(user_id, weibo_id) VALUES(:user_id, :weibo_id)", comment)
	if err != nil {
		return err
	}
	return err
}

func (wb *WeiboRepository) AddCommentNumByWeiboID(weiboID int64, num int32) error {
	_, err := wb.db.Exec("UPDATE `weibos` SET Comment_num = weibo_id_num + ? WHERE id = ?", num, weiboID)
	return err
}

func (wb *WeiboRepository) GetCommentByID(commentID int64) (*weibo.Comment, error) {
	var weibo weibo.Comment
	if err := wb.db.Get(&weibo, "SELECT * FROM `comment` WHERE `weiboID` = ?", commentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &weibo, nil
}

// 删除收藏
func (wb *WeiboRepository) DeleteComment(weiboID int64) error {
	_, err := wb.db.Exec("UPDATE `comment` SET weibo_id = weibo_id + ? WHERE id = ?", weiboID)
	return err
}

func (wb *WeiboRepository) GetWeibosByUserTimelines(userID int64, offset, limit int64) ([]*weibo.WeiboWithUser, error) {
	weibos := []*weibo.WeiboWithUser{}
	key := fmt.Sprintf("weibos:%d", userID)

	data, err := redis.Bytes(wb.redisClient.Do("GET", key))
	if err == nil {
		if err = json.Unmarshal(data, &weibos); err != nil {
			return nil, err
		}
		return weibos, nil
	}

	if err != redis.ErrNil {
		return nil, err
	}

	query := `
		SELECT w.*, u.avatar FROM weibos w
		INNER JOIN timeline t ON w.id = t.weibo_id AND t.user_id = ?
		INNER JOIN users u ON w.user_id = u.id
		ORDER BY t.weibo_created_at DESC LIMIT ?, ?
	`
	if err := wb.db.Select(&weibos, query, userID, offset, limit); err != nil {
		return nil, err
	}

	data, err = json.Marshal(weibos)
	if err != nil {
		return nil, err
	}

	_, err = wb.redisClient.Do("SETEX", key, 600, data)
	if err != nil {
		return nil, err
	}

	return weibos, nil
}

func (wb *WeiboRepository) GetWeibosByAccountOrContent(accountOrContent string, offset, perPage int64) ([]*weibo.Weibo, error) {
	weibos := []*weibo.Weibo{}
	if err := wb.db.Select(&weibos, "SELECT * FROM `weibos` WHERE `account` = ? OR `content` LIKE ? ORDER BY created_at DESC LIMIT ?, ?",
		accountOrContent, "%"+accountOrContent+"%", offset, perPage); err != nil {
		return nil, err
	}
	return weibos, nil
}
