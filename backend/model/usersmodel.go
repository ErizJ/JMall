package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UsersModel = (*customUsersModel)(nil)

type (
	// UsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersModel.
	UsersModel interface {
		usersModel
		withSession(session sqlx.Session) UsersModel
		FindAll(ctx context.Context) ([]*Users, error)
		FindAllPaged(ctx context.Context, page, pageSize int64) ([]*Users, int64, error)
	}

	customUsersModel struct {
		*defaultUsersModel
	}
)

// NewUsersModel returns a model for the database table.
func NewUsersModel(conn sqlx.SqlConn) UsersModel {
	return &customUsersModel{
		defaultUsersModel: newUsersModel(conn),
	}
}

func (m *customUsersModel) withSession(session sqlx.Session) UsersModel {
	return NewUsersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUsersModel) FindAll(ctx context.Context) ([]*Users, error) {
	query := fmt.Sprintf("select %s from %s", usersRows, m.table)
	var resp []*Users
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customUsersModel) FindAllPaged(ctx context.Context, page, pageSize int64) ([]*Users, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s", m.table)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("select %s from %s limit ? offset ?", usersRows, m.table)
	var resp []*Users
	err := m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}
