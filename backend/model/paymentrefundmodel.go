package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 退款状态常量
const (
	RefundStatusPending = 0 // 退款中
	RefundStatusSuccess = 1 // 退款成功
	RefundStatusFailed  = 2 // 退款失败
)

var _ PaymentRefundModel = (*customPaymentRefundModel)(nil)

type (
	PaymentRefundModel interface {
		paymentRefundModel
		WithSession(session sqlx.Session) PaymentRefundModel
		FindByRefundNo(ctx context.Context, refundNo string) (*PaymentRefund, error)
		FindByPaymentNo(ctx context.Context, paymentNo string) ([]*PaymentRefund, error)
		UpdateStatus(ctx context.Context, refundNo string, status int64, updatedAt int64) error
	}

	customPaymentRefundModel struct {
		*defaultPaymentRefundModel
	}
)

func NewPaymentRefundModel(conn sqlx.SqlConn) PaymentRefundModel {
	return &customPaymentRefundModel{
		defaultPaymentRefundModel: newPaymentRefundModel(conn),
	}
}

func (m *customPaymentRefundModel) WithSession(session sqlx.Session) PaymentRefundModel {
	return NewPaymentRefundModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customPaymentRefundModel) FindByRefundNo(ctx context.Context, refundNo string) (*PaymentRefund, error) {
	query := fmt.Sprintf("select %s from %s where `refund_no` = ? limit 1", paymentRefundRows, m.table)
	var resp PaymentRefund
	err := m.conn.QueryRowCtx(ctx, &resp, query, refundNo)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customPaymentRefundModel) FindByPaymentNo(ctx context.Context, paymentNo string) ([]*PaymentRefund, error) {
	query := fmt.Sprintf("select %s from %s where `payment_no` = ? order by `created_at` desc", paymentRefundRows, m.table)
	var resp []*PaymentRefund
	err := m.conn.QueryRowsCtx(ctx, &resp, query, paymentNo)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customPaymentRefundModel) UpdateStatus(ctx context.Context, refundNo string, status int64, updatedAt int64) error {
	query := fmt.Sprintf("update %s set `status` = ?, `updated_at` = ? where `refund_no` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, updatedAt, refundNo)
	return err
}
