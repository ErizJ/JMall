package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// OrderDetail is a flattened view of an order joined with user and product info.
type OrderDetail struct {
	Orders
	UserName    string `db:"user_name"`
	ProductName string `db:"product_name"`
	ProductImg  string `db:"product_picture"`
}

var _ OrdersModel = (*customOrdersModel)(nil)

type (
	// OrdersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrdersModel.
	OrdersModel interface {
		ordersModel
		withSession(session sqlx.Session) OrdersModel
		WithSession(session sqlx.Session) OrdersModel
		TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error
		FindByUserId(ctx context.Context, userId int64) ([]*Orders, error)
		FindByOrderId(ctx context.Context, orderId int64) ([]*Orders, error)
		DeleteByOrderId(ctx context.Context, orderId int64) error
		FindAll(ctx context.Context) ([]*Orders, error)
		FindByUserIdGrouped(ctx context.Context, userId int64) ([]int64, error)
		FindAllWithDetails(ctx context.Context) ([]*OrderDetail, error)
		UpdateStatusByOrderId(ctx context.Context, orderId int64, status int64) error
	}

	customOrdersModel struct {
		*defaultOrdersModel
	}
)

// NewOrdersModel returns a model for the database table.
func NewOrdersModel(conn sqlx.SqlConn) OrdersModel {
	return &customOrdersModel{
		defaultOrdersModel: newOrdersModel(conn),
	}
}

func (m *customOrdersModel) withSession(session sqlx.Session) OrdersModel {
	return NewOrdersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customOrdersModel) WithSession(session sqlx.Session) OrdersModel {
	return NewOrdersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customOrdersModel) TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}

func (m *customOrdersModel) FindByUserId(ctx context.Context, userId int64) ([]*Orders, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=? order by `order_time` desc", ordersRows, m.table)
	var resp []*Orders
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customOrdersModel) FindByOrderId(ctx context.Context, orderId int64) ([]*Orders, error) {
	query := fmt.Sprintf("select %s from %s where `order_id`=?", ordersRows, m.table)
	var resp []*Orders
	err := m.conn.QueryRowsCtx(ctx, &resp, query, orderId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customOrdersModel) DeleteByOrderId(ctx context.Context, orderId int64) error {
	query := fmt.Sprintf("delete from %s where `order_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, orderId)
	return err
}

func (m *customOrdersModel) FindAll(ctx context.Context) ([]*Orders, error) {
	query := fmt.Sprintf("select %s from %s", ordersRows, m.table)
	var resp []*Orders
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customOrdersModel) FindByUserIdGrouped(ctx context.Context, userId int64) ([]int64, error) {
	query := fmt.Sprintf("select distinct `order_id` from %s where `user_id`=? order by `order_id` desc", m.table)
	var resp []int64
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customOrdersModel) FindAllWithDetails(ctx context.Context) ([]*OrderDetail, error) {
	query := `select o.id, o.order_id, o.user_id, o.product_id, o.product_num, o.product_price, o.order_time, o.status,
		u.user_name, p.product_name, COALESCE(p.product_picture, '') as product_picture
		from orders o
		join users u on o.user_id = u.user_id
		join product p on o.product_id = p.product_id
		order by o.order_time desc`
	var resp []*OrderDetail
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateStatusByOrderId 更新订单状态（支付回调联动用）
// 0=待支付 1=已支付 2=已取消 3=已退款
func (m *customOrdersModel) UpdateStatusByOrderId(ctx context.Context, orderId int64, status int64) error {
	query := fmt.Sprintf("update %s set `status` = ? where `order_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, orderId)
	return err
}
