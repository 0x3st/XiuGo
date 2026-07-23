// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsSessionDao is the data access object for the table bbs_session.
type BbsSessionDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsSessionColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsSessionColumns defines and stores column names for the table bbs_session.
type BbsSessionColumns struct {
	Sid       string //
	Uid       string //
	Fid       string //
	Url       string //
	Ip        string //
	Useragent string //
	Data      string //
	Bigdata   string //
	LastDate  string //
}

// bbsSessionColumns holds the columns for the table bbs_session.
var bbsSessionColumns = BbsSessionColumns{
	Sid:       "sid",
	Uid:       "uid",
	Fid:       "fid",
	Url:       "url",
	Ip:        "ip",
	Useragent: "useragent",
	Data:      "data",
	Bigdata:   "bigdata",
	LastDate:  "last_date",
}

// NewBbsSessionDao creates and returns a new DAO object for table data access.
func NewBbsSessionDao(handlers ...gdb.ModelHandler) *BbsSessionDao {
	return &BbsSessionDao{
		group:    "default",
		table:    "bbs_session",
		columns:  bbsSessionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsSessionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsSessionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsSessionDao) Columns() BbsSessionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsSessionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsSessionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsSessionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
