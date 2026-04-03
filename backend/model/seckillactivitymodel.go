package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SeckillActivityModel = (*customSeckillActivityModel)(nil)

type (
	SeckillActivityModel interface {
		seckillActivityModel
		WithSession(session sqlx.Session) SeckillActivityModel
		FindActiveByProductId(ctx context.Context, productId int64) (*SeckillActivity, error)
		FindUpcoming(ctx context.Context, now int64) ([]*SeckillActivity, error)
		FindOngoing(ctx context.Context, now int64) ([]*SeckillActivity, error)
		DecrStock(ctx context.Context, id int64, num int64) error
		IncrStock(ctx context.Context, id int64, num int64) error
		UpdateStatus(ctx context.Context, id int64, status int64) error
	}

	customSeckillActivityModel struct {
		*defaultSeckillActivityModel
	}
)

func NewSeckillActivityModel(conn sqlx.SqlConn) SeckillActivityModel {
	return &customSeckillActivityModel{
		defaultSeckillActivityModel: newSeckillActivityModel(conn),
	}
}

func (m *customSeckillActivityModel) WithSession(session sqlx.Session) SeckillActivityModel {
	return NewSeckillActivityModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customSeckillActivityModel) FindActiveByProductId(ctx context.Context, productId int64) (*SeckillActivity, error) {
	query := fmt.Sprintf("select %s from %s where `product_id` = ? and `status` = 1 limit 1", seckillActivityRows, m.table)
	var resp SeckillActivity
	err := m.conn.QueryRowCtx(ctx, &resp, query, productId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindUpcoming returns activities starting within the next hour.
func (m *customSeckillActivityModel) FindUpcoming(ctx context.Context, now int64) ([]*SeckillActivity, error) {
	query := fmt.Sprintf("select %s from %s where `status` = 0 and `start_time` <= ? order by `start_time` asc", seckillActivityRows, m.table)
	var resp []*SeckillActivity
	err := m.conn.QueryRowsCtx(ctx, &resp, query, now+3600)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindOngoing returns all currently active seckill activities.
func (m *customSeckillActivityModel) FindOngoing(ctx context.Context, now int64) ([]*SeckillActivity, error) {
	query := fmt.Sprintf("select %s from %s where `status` = 1 and `start_time` <= ? and `end_time` >= ?", seckillActivityRows, m.table)
	var resp []*SeckillActivity
	err := m.conn.QueryRowsCtx(ctx, &resp, query, now, now)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DecrStock atomically decrements available_stock. Returns error if stock insufficient.
func (m *customSeckillActivityModel) DecrStock(ctx context.Context, id int64, num int64) error {
	query := fmt.Sprintf("update %s set `available_stock` = `available_stock` - ? where `id` = ? and `available_stock` >= ?", m.table)
	result, err := m.conn.ExecCtx(ctx, query, num, id, num)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("seckill stock insufficient for activity %d", id)
	}
	return nil
}

// IncrStock rolls back stock (used when order creation fails).
func (m *customSeckillActivityModel) IncrStock(ctx context.Context, id int64, num int64) error {
	query := fmt.Sprintf("update %s set `available_stock` = `available_stock` + ? where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, num, id)
	return err
}

// UpdateStatus updates the activity status.
func (m *customSeckillActivityModel) UpdateStatus(ctx context.Context, id int64, status int64) error {
	query := fmt.Sprintf("update %s set `status` = ? where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}
