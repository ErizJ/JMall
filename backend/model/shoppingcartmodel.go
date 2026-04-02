package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ShoppingcartModel = (*customShoppingcartModel)(nil)

type (
	// ShoppingcartModel is an interface to be customized, add more methods here,
	// and implement the added methods in customShoppingcartModel.
	ShoppingcartModel interface {
		shoppingcartModel
		withSession(session sqlx.Session) ShoppingcartModel
		WithSession(session sqlx.Session) ShoppingcartModel
		FindByUserId(ctx context.Context, userId int64) ([]*Shoppingcart, error)
		FindByUserAndProduct(ctx context.Context, userId, productId int64) (*Shoppingcart, error)
		UpdateNumByUserAndProduct(ctx context.Context, userId, productId, num int64) error
		DeleteByUserAndProduct(ctx context.Context, userId, productId int64) error
		DeleteByUserId(ctx context.Context, userId int64) error
	}

	customShoppingcartModel struct {
		*defaultShoppingcartModel
	}
)

// NewShoppingcartModel returns a model for the database table.
func NewShoppingcartModel(conn sqlx.SqlConn) ShoppingcartModel {
	return &customShoppingcartModel{
		defaultShoppingcartModel: newShoppingcartModel(conn),
	}
}

func (m *customShoppingcartModel) withSession(session sqlx.Session) ShoppingcartModel {
	return NewShoppingcartModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customShoppingcartModel) WithSession(session sqlx.Session) ShoppingcartModel {
	return NewShoppingcartModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customShoppingcartModel) FindByUserId(ctx context.Context, userId int64) ([]*Shoppingcart, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=?", shoppingcartRows, m.table)
	var resp []*Shoppingcart
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customShoppingcartModel) FindByUserAndProduct(ctx context.Context, userId, productId int64) (*Shoppingcart, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=? and `product_id`=? limit 1", shoppingcartRows, m.table)
	var resp Shoppingcart
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

func (m *customShoppingcartModel) UpdateNumByUserAndProduct(ctx context.Context, userId, productId, num int64) error {
	query := fmt.Sprintf("update %s set `num`=? where `user_id`=? and `product_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, num, userId, productId)
	return err
}

func (m *customShoppingcartModel) DeleteByUserAndProduct(ctx context.Context, userId, productId int64) error {
	query := fmt.Sprintf("delete from %s where `user_id`=? and `product_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, productId)
	return err
}

func (m *customShoppingcartModel) DeleteByUserId(ctx context.Context, userId int64) error {
	query := fmt.Sprintf("delete from %s where `user_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}
