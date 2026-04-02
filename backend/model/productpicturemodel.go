package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ProductPictureModel = (*customProductPictureModel)(nil)

type (
	// ProductPictureModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductPictureModel.
	ProductPictureModel interface {
		productPictureModel
		withSession(session sqlx.Session) ProductPictureModel
		FindByProductId(ctx context.Context, productId int64) ([]*ProductPicture, error)
	}

	customProductPictureModel struct {
		*defaultProductPictureModel
	}
)

// NewProductPictureModel returns a model for the database table.
func NewProductPictureModel(conn sqlx.SqlConn) ProductPictureModel {
	return &customProductPictureModel{
		defaultProductPictureModel: newProductPictureModel(conn),
	}
}

func (m *customProductPictureModel) withSession(session sqlx.Session) ProductPictureModel {
	return NewProductPictureModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customProductPictureModel) FindByProductId(ctx context.Context, productId int64) ([]*ProductPicture, error) {
	query := fmt.Sprintf("select %s from %s where `product_id`=?", productPictureRows, m.table)
	var resp []*ProductPicture
	err := m.conn.QueryRowsCtx(ctx, &resp, query, productId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
