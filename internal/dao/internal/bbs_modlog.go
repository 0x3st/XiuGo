// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsModlogDao is the data access object for the table bbs_modlog.
type BbsModlogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsModlogColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsModlogColumns defines and stores column names for the table bbs_modlog.
type BbsModlogColumns struct {
	Logid      string //
	Uid        string //
	Tid        string //
	Pid        string //
	Subject    string //
	Comment    string //
	Rmbs       string //
	CreateDate string //
	Action     string //
}

// bbsModlogColumns holds the columns for the table bbs_modlog.
var bbsModlogColumns = BbsModlogColumns{
	Logid:      "logid",
	Uid:        "uid",
	Tid:        "tid",
	Pid:        "pid",
	Subject:    "subject",
	Comment:    "comment",
	Rmbs:       "rmbs",
	CreateDate: "create_date",
	Action:     "action",
}

// NewBbsModlogDao creates and returns a new DAO object for table data access.
func NewBbsModlogDao(handlers ...gdb.ModelHandler) *BbsModlogDao {
	return &BbsModlogDao{
		group:    "default",
		table:    "bbs_modlog",
		columns:  bbsModlogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsModlogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsModlogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsModlogDao) Columns() BbsModlogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsModlogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsModlogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *BbsModlogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
