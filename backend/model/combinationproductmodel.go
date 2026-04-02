package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CombinationProductModel = (*customCombinationProductModel)(nil)

type (
	// CombinationProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCombinationProductModel.
	CombinationProductModel interface {
		combinationProductModel
		withSession(session sqlx.Session) CombinationProductModel
		FindAll(ctx context.Context) ([]*CombinationProduct, error)
		FindByMainProductId(ctx context.Context, mainProductId int64) ([]*CombinationProduct, error)
	}

	customCombinationProductModel struct {
		*defaultCombinationProductModel
	}
)

// NewCombinationProductModel returns a model for the database table.
func NewCombinationProductModel(conn sqlx.SqlConn) CombinationProductModel {
	return &customCombinationProductModel{
		defaultCombinationProductModel: newCombinationProductModel(conn),
	}
}

func (m *customCombinationProductModel) withSession(session sqlx.Session) CombinationProductModel {
	return NewCombinationProductModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customCombinationProductModel) FindAll(ctx context.Context) ([]*CombinationProduct, error) {
	query := fmt.Sprintf("select %s from %s", combinationProductRows, m.table)
	var resp []*CombinationProduct
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customCombinationProductModel) FindByMainProductId(ctx context.Context, mainProductId int64) ([]*CombinationProduct, error) {
	query := fmt.Sprintf("select %s from %s where `main_product_id`=?", combinationProductRows, m.table)
	var resp []*CombinationProduct
	err := m.conn.QueryRowsCtx(ctx, &resp, query, mainProductId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
