package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 支付状态常量
const (
	PaymentStatusPending = 0 // 待支付
	PaymentStatusPaying  = 1 // 支付中
	PaymentStatusSuccess = 2 // 支付成功
	PaymentStatusFailed  = 3 // 支付失败
	PaymentStatusClosed  = 4 // 已关闭
	PaymentStatusRefund  = 5 // 已退款
)

var _ PaymentOrderModel = (*customPaymentOrderModel)(nil)

type (
	PaymentOrderModel interface {
		paymentOrderModel
		WithSession(session sqlx.Session) PaymentOrderModel
		TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error
		FindByPaymentNo(ctx context.Context, paymentNo string) (*PaymentOrder, error)
		FindByOrderId(ctx context.Context, orderId int64) ([]*PaymentOrder, error)
		FindByUserId(ctx context.Context, userId int64) ([]*PaymentOrder, error)
		UpdateStatus(ctx context.Context, paymentNo string, status int64, updatedAt int64) error
		UpdatePaySuccess(ctx context.Context, paymentNo string, channelTradeNo string, paidTime int64, updatedAt int64) error
	}

	customPaymentOrderModel struct {
		*defaultPaymentOrderModel
	}
)

func NewPaymentOrderModel(conn sqlx.SqlConn) PaymentOrderModel {
	return &customPaymentOrderModel{
		defaultPaymentOrderModel: newPaymentOrderModel(conn),
	}
}

func (m *customPaymentOrderModel) WithSession(session sqlx.Session) PaymentOrderModel {
	return NewPaymentOrderModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customPaymentOrderModel) TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}

func (m *customPaymentOrderModel) FindByPaymentNo(ctx context.Context, paymentNo string) (*PaymentOrder, error) {
	query := fmt.Sprintf("select %s from %s where `payment_no` = ? limit 1", paymentOrderRows, m.table)
	var resp PaymentOrder
	err := m.conn.QueryRowCtx(ctx, &resp, query, paymentNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customPaymentOrderModel) FindByOrderId(ctx context.Context, orderId int64) ([]*PaymentOrder, error) {
	query := fmt.Sprintf("select %s from %s where `order_id` = ? order by `created_at` desc", paymentOrderRows, m.table)
	var resp []*PaymentOrder
	err := m.conn.QueryRowsCtx(ctx, &resp, query, orderId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customPaymentOrderModel) FindByUserId(ctx context.Context, userId int64) ([]*PaymentOrder, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc", paymentOrderRows, m.table)
	var resp []*PaymentOrder
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateStatus 更新支付单状态（通用）
func (m *customPaymentOrderModel) UpdateStatus(ctx context.Context, paymentNo string, status int64, updatedAt int64) error {
	query := fmt.Sprintf("update %s set `status` = ?, `updated_at` = ? where `payment_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, updatedAt, paymentNo)
	return err
}

// UpdatePaySuccess 支付成功时更新（原子操作：状态+渠道交易号+支付时间）
// WHERE 条件加了 status=0 OR status=1，防止重复更新已成功的单
func (m *customPaymentOrderModel) UpdatePaySuccess(ctx context.Context, paymentNo string, channelTradeNo string, paidTime int64, updatedAt int64) error {
	query := fmt.Sprintf("update %s set `status` = ?, `channel_trade_no` = ?, `paid_time` = ?, `updated_at` = ? where `payment_no` = ? and `status` in (0, 1)", m.table)
	_, err := m.conn.ExecCtx(ctx, query, PaymentStatusSuccess, channelTradeNo, paidTime, updatedAt, paymentNo)
	return err
}
