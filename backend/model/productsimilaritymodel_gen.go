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
	productSimilarityFieldNames          = builder.RawFieldNames(&ProductSimilarity{})
	productSimilarityRows                = strings.Join(productSimilarityFieldNames, ",")
	productSimilarityRowsExpectAutoSet   = strings.Join(stringx.Remove(productSimilarityFieldNames, "`id`"), ",")
	productSimilarityRowsWithPlaceHolder = strings.Join(stringx.Remove(productSimilarityFieldNames, "`id`"), "=?,") + "=?"
)

type (
	productSimilarityModel interface {
		Insert(ctx context.Context, data *ProductSimilarity) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ProductSimilarity, error)
		Delete(ctx context.Context, id int64) error
	}

	defaultProductSimilarityModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ProductSimilarity struct {
		Id               int64   `db:"id"`
		ProductId        int64   `db:"product_id"`
		SimilarProductId int64   `db:"similar_product_id"`
		Score            float64 `db:"score"`
		UpdatedAt        int64   `db:"updated_at"`
	}
)

func newProductSimilarityModel(conn sqlx.SqlConn) *defaultProductSimilarityModel {
	return &defaultProductSimilarityModel{conn: conn, table: "`product_similarity`"}
}

func (m *defaultProductSimilarityModel) Insert(ctx context.Context, data *ProductSimilarity) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?)", m.table, productSimilarityRowsExpectAutoSet)
	return m.conn.ExecCtx(ctx, query, data.ProductId, data.SimilarProductId, data.Score, data.UpdatedAt)
}

func (m *defaultProductSimilarityModel) FindOne(ctx context.Context, id int64) (*ProductSimilarity, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", productSimilarityRows, m.table)
	var resp ProductSimilarity
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

func (m *defaultProductSimilarityModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultProductSimilarityModel) tableName() string {
	return m.table
}
