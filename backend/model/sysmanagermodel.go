package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SysmanagerModel = (*customSysmanagerModel)(nil)

type (
	// SysmanagerModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysmanagerModel.
	SysmanagerModel interface {
		sysmanagerModel
		withSession(session sqlx.Session) SysmanagerModel
	}

	customSysmanagerModel struct {
		*defaultSysmanagerModel
	}
)

// NewSysmanagerModel returns a model for the database table.
func NewSysmanagerModel(conn sqlx.SqlConn) SysmanagerModel {
	return &customSysmanagerModel{
		defaultSysmanagerModel: newSysmanagerModel(conn),
	}
}

func (m *customSysmanagerModel) withSession(session sqlx.Session) SysmanagerModel {
	return NewSysmanagerModel(sqlx.NewSqlConnFromSession(session))
}
