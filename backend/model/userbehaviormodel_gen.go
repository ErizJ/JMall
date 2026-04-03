package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	userBehaviorFieldNames          = builder.RawFieldNames(&UserBehavior{})
	userBehaviorRows                = strings.Join(userBehaviorFieldNames, ",")
	userBehaviorRowsExpectAutoSet   = strings.Join(stringx.Remove(userBehaviorFieldNames, "`id`"), ",")
	userBehaviorRowsWithPlaceHolder = strings.Join(stringx.Remove(userBehaviorFieldNames, "`id`"), "=?,") + "=?"
)

type (
	userBehaviorModel interface {
		Insert(ctx context.Context, data *UserBehavior) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*UserBehavior, error)
		Delete(ctx context.Context, id int64) error
	}

	defaultUserBehaviorModel struct {
		conn  sqlx.SqlConn
		table string
	}

	UserBehavior struct {
		Id           int64 `db:"id"`
		UserId       int64 `db:"user_id"`
		ProductId    int64 `db:"product_id"`
		CategoryId   int64 `db:"category_id"`
		BehaviorType int64 `db:"behavior_type"` // 1=浏览 2=点击 3=加购 4=购买 5=收藏
		BehaviorTime int64 `db:"behavior_time"`
	}
)

func newUserBehaviorModel(conn sqlx.SqlConn) *defaultUserBehaviorModel {
	return &defaultUserBehaviorModel{conn: conn, table: "`user_behavior`"}
}

func (m *defaultUserBehaviorModel) Insert(ctx context.Context, data *UserBehavior) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?)", m.table, userBehaviorRowsExpectAutoSet)
	return m.conn.ExecCtx(ctx, query, data.UserId, data.ProductId, data.CategoryId, data.BehaviorType, data.BehaviorTime)
}

func (m *defaultUserBehaviorModel) FindOne(ctx context.Context, id int64) (*UserBehavior, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userBehaviorRows, m.table)
	var resp UserBehavior
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultUserBehaviorModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultUserBehaviorModel) tableName() string {
	return m.table
}
