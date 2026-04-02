package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CollectModel = (*customCollectModel)(nil)

type (
	// CollectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCollectModel.
	CollectModel interface {
		collectModel
		withSession(session sqlx.Session) CollectModel
		FindByUserId(ctx context.Context, userId int64) ([]*Collect, error)
		FindByUserAndProduct(ctx context.Context, userId, productId int64) (*Collect, error)
		DeleteByUserAndProduct(ctx context.Context, userId, productId int64) error
	}

	customCollectModel struct {
		*defaultCollectModel
	}
)

// NewCollectModel returns a model for the database table.
func NewCollectModel(conn sqlx.SqlConn) CollectModel {
	return &customCollectModel{
		defaultCollectModel: newCollectModel(conn),
	}
}

func (m *customCollectModel) withSession(session sqlx.Session) CollectModel {
	return NewCollectModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customCollectModel) FindByUserId(ctx context.Context, userId int64) ([]*Collect, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=?", collectRows, m.table)
	var resp []*Collect
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customCollectModel) FindByUserAndProduct(ctx context.Context, userId, productId int64) (*Collect, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=? and `product_id`=? limit 1", collectRows, m.table)
	var resp Collect
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, productId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customCollectModel) DeleteByUserAndProduct(ctx context.Context, userId, productId int64) error {
	query := fmt.Sprintf("delete from %s where `user_id`=? and `product_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, productId)
	return err
}
