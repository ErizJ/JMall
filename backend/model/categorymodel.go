package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CategoryModel = (*customCategoryModel)(nil)

type (
	// CategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCategoryModel.
	CategoryModel interface {
		categoryModel
		withSession(session sqlx.Session) CategoryModel
		FindAll(ctx context.Context) ([]*Category, error)
		FindOneByCategoryName(ctx context.Context, name string) (*Category, error)
		IncrCategoryHot(ctx context.Context, categoryId int64) error
		ResetAllCategoryHot(ctx context.Context) error
	}

	customCategoryModel struct {
		*defaultCategoryModel
	}
)

// NewCategoryModel returns a model for the database table.
func NewCategoryModel(conn sqlx.SqlConn) CategoryModel {
	return &customCategoryModel{
		defaultCategoryModel: newCategoryModel(conn),
	}
}

func (m *customCategoryModel) withSession(session sqlx.Session) CategoryModel {
	return NewCategoryModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customCategoryModel) FindAll(ctx context.Context) ([]*Category, error) {
	query := fmt.Sprintf("select %s from %s", categoryRows, m.table)
	var resp []*Category
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customCategoryModel) FindOneByCategoryName(ctx context.Context, name string) (*Category, error) {
	query := fmt.Sprintf("select %s from %s where `category_name` = ? limit 1", categoryRows, m.table)
	var resp Category
	err := m.conn.QueryRowCtx(ctx, &resp, query, name)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customCategoryModel) IncrCategoryHot(ctx context.Context, categoryId int64) error {
	query := fmt.Sprintf("update %s set `category_hot`=`category_hot`+1 where `category_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, categoryId)
	return err
}

func (m *customCategoryModel) ResetAllCategoryHot(ctx context.Context) error {
	query := fmt.Sprintf("update %s set `category_hot`=0", m.table)
	_, err := m.conn.ExecCtx(ctx, query)
	return err
}
