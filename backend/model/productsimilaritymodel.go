package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ProductSimilarityModel = (*customProductSimilarityModel)(nil)

type (
	ProductSimilarityModel interface {
		productSimilarityModel
		// 查询某商品的相似商品列表（按相似度降序）
		FindSimilarProducts(ctx context.Context, productId int64, limit int) ([]*ProductSimilarity, error)
		// 批量查询多个商品的相似商品
		FindSimilarProductsByIds(ctx context.Context, productIds []int64, limit int) ([]*ProductSimilarity, error)
		// 批量写入/更新相似度（UPSERT）
		BatchUpsert(ctx context.Context, items []*ProductSimilarity) error
	}

	customProductSimilarityModel struct {
		*defaultProductSimilarityModel
	}
)

func NewProductSimilarityModel(conn sqlx.SqlConn) ProductSimilarityModel {
	return &customProductSimilarityModel{
		defaultProductSimilarityModel: newProductSimilarityModel(conn),
	}
}

func (m *customProductSimilarityModel) FindSimilarProducts(ctx context.Context, productId int64, limit int) ([]*ProductSimilarity, error) {
	query := fmt.Sprintf("select %s from %s where `product_id`=? order by `score` desc limit ?", productSimilarityRows, m.table)
	var resp []*ProductSimilarity
	err := m.conn.QueryRowsCtx(ctx, &resp, query, productId, limit)
	return resp, err
}

func (m *customProductSimilarityModel) FindSimilarProductsByIds(ctx context.Context, productIds []int64, limit int) ([]*ProductSimilarity, error) {
	if len(productIds) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(productIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf("select %s from %s where `product_id` in (%s) order by `score` desc limit ?",
		productSimilarityRows, m.table, placeholders)
	args := make([]interface{}, 0, len(productIds)+1)
	for _, id := range productIds {
		args = append(args, id)
	}
	args = append(args, limit)
	var resp []*ProductSimilarity
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customProductSimilarityModel) BatchUpsert(ctx context.Context, items []*ProductSimilarity) error {
	if len(items) == 0 {
		return nil
	}
	values := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*4)
	for _, item := range items {
		values = append(values, "(?, ?, ?, ?)")
		args = append(args, item.ProductId, item.SimilarProductId, item.Score, item.UpdatedAt)
	}
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE `score`=VALUES(`score`), `updated_at`=VALUES(`updated_at`)",
		m.table, productSimilarityRowsExpectAutoSet, strings.Join(values, ","))
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}
