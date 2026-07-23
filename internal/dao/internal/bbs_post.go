// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsPostDao is the data access object for the table bbs_post.
type BbsPostDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsPostColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsPostColumns defines and stores column names for the table bbs_post.
type BbsPostColumns struct {
	Tid        string //
	Pid        string //
	Uid        string //
	Isfirst    string //
	CreateDate string //
	Userip     string //
	Images     string //
	Files      string //
	Doctype    string //
	Quotepid   string //
	Message    string //
	MessageFmt string //
}

// bbsPostColumns holds the columns for the table bbs_post.
var bbsPostColumns = BbsPostColumns{
	Tid:        "tid",
	Pid:        "pid",
	Uid:        "uid",
	Isfirst:    "isfirst",
	CreateDate: "create_date",
	Userip:     "userip",
	Images:     "images",
	Files:      "files",
	Doctype:    "doctype",
	Quotepid:   "quotepid",
	Message:    "message",
	MessageFmt: "message_fmt",
}

// NewBbsPostDao creates and returns a new DAO object for table data access.
func NewBbsPostDao(handlers ...gdb.ModelHandler) *BbsPostDao {
	return &BbsPostDao{
		group:    "default",
		table:    "bbs_post",
		columns:  bbsPostColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsPostDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsPostDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsPostDao) Columns() BbsPostColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsPostDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsPostDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsPostDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
