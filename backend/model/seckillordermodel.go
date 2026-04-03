package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SeckillOrderModel = (*customSeckillOrderModel)(nil)

type (
	SeckillOrderModel interface {
		seckillOrderModel
		WithSession(session sqlx.Session) SeckillOrderModel
		TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error
		FindByActivityAndUser(ctx context.Context, activityId, userId int64) (*SeckillOrder, error)
		FindByUserId(ctx context.Context, userId int64) ([]*SeckillOrder, error)
	}

	customSeckillOrderModel struct {
		*defaultSeckillOrderModel
	}
)

func NewSeckillOrderModel(conn sqlx.SqlConn) SeckillOrderModel {
	return &customSeckillOrderModel{
		defaultSeckillOrderModel: newSeckillOrderModel(conn),
	}
}

func (m *customSeckillOrderModel) WithSession(session sqlx.Session) SeckillOrderModel {
	return NewSeckillOrderModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customSeckillOrderModel) TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}

func (m *customSeckillOrderModel) FindByActivityAndUser(ctx context.Context, activityId, userId int64) (*SeckillOrder, error) {
	query := fmt.Sprintf("select %s from %s where `activity_id` = ? and `user_id` = ? limit 1", seckillOrderRows, m.table)
	var resp SeckillOrder
	err := m.conn.QueryRowCtx(ctx, &resp, query, activityId, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByUserId returns all seckill orders for a user.
func (m *customSeckillOrderModel) FindByUserId(ctx context.Context, userId int64) ([]*SeckillOrder, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc", seckillOrderRows, m.table)
	var resp []*SeckillOrder
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
