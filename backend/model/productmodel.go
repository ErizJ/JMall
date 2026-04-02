package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ProductModel = (*customProductModel)(nil)

type (
	// ProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductModel.
	ProductModel interface {
		productModel
		withSession(session sqlx.Session) ProductModel
		FindAll(ctx context.Context) ([]*Product, error)
		FindByIds(ctx context.Context, ids []int64) ([]*Product, error)
		FindByCategory(ctx context.Context, categoryId int64) ([]*Product, error)
		FindBySearch(ctx context.Context, keyword string) ([]*Product, error)
		FindTopHot(ctx context.Context, limit int) ([]*Product, error)
		FindTopHotByCategory(ctx context.Context, categoryId int64, limit int) ([]*Product, error)
		FindByIsPromotion(ctx context.Context, limit int) ([]*Product, error)
		IncrProductHot(ctx context.Context, productId int64) error
	}

	customProductModel struct {
		*defaultProductModel
	}
)

// NewProductModel returns a model for the database table.
func NewProductModel(conn sqlx.SqlConn) ProductModel {
	return &customProductModel{
		defaultProductModel: newProductModel(conn),
	}
}

func (m *customProductModel) withSession(session sqlx.Session) ProductModel {
	return NewProductModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customProductModel) FindAll(ctx context.Context) ([]*Product, error) {
	query := fmt.Sprintf("select %s from %s", productRows, m.table)
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customProductModel) FindByCategory(ctx context.Context, categoryId int64) ([]*Product, error) {
	query := fmt.Sprintf("select %s from %s where `category_id`=? order by `product_sales` desc", productRows, m.table)
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query, categoryId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customProductModel) FindBySearch(ctx context.Context, keyword string) ([]*Product, error) {
	like := "%" + keyword + "%"
	query := fmt.Sprintf("select %s from %s where `product_name` like ? or `product_title` like ? or `product_intro` like ?", productRows, m.table)
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query, like, like, like)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customProductModel) FindTopHot(ctx context.Context, limit int) ([]*Product, error) {
	query := fmt.Sprintf("select %s from %s order by `product_hot` desc limit ?", productRows, m.table)
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customProductModel) FindTopHotByCategory(ctx context.Context, categoryId int64, limit int) ([]*Product, error) {
	query := fmt.Sprintf("select %s from %s where `category_id`=? order by `product_hot` desc limit ?", productRows, m.table)
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query, categoryId, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customProductModel) FindByIsPromotion(ctx context.Context, limit int) ([]*Product, error) {
	query := fmt.Sprintf("select %s from %s where `product_isPromotion` > 0 order by `product_isPromotion` desc limit ?", productRows, m.table)
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customProductModel) IncrProductHot(ctx context.Context, productId int64) error {
	query := fmt.Sprintf("update %s set `product_hot`=`product_hot`+1 where `product_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, productId)
	return err
}

func (m *customProductModel) FindByIds(ctx context.Context, ids []int64) ([]*Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("select %s from %s where `product_id` in (%s)", productRows, m.table, placeholders)
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	var resp []*Product
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
