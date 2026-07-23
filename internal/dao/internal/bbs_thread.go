// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsThreadDao is the data access object for the table bbs_thread.
type BbsThreadDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsThreadColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsThreadColumns defines and stores column names for the table bbs_thread.
type BbsThreadColumns struct {
	Fid        string //
	Tid        string //
	Top        string //
	Uid        string //
	Userip     string //
	Subject    string //
	CreateDate string //
	LastDate   string //
	Views      string //
	Posts      string //
	Images     string //
	Files      string //
	Mods       string //
	Closed     string //
	Firstpid   string //
	Lastuid    string //
	Lastpid    string //
}

// bbsThreadColumns holds the columns for the table bbs_thread.
var bbsThreadColumns = BbsThreadColumns{
	Fid:        "fid",
	Tid:        "tid",
	Top:        "top",
	Uid:        "uid",
	Userip:     "userip",
	Subject:    "subject",
	CreateDate: "create_date",
	LastDate:   "last_date",
	Views:      "views",
	Posts:      "posts",
	Images:     "images",
	Files:      "files",
	Mods:       "mods",
	Closed:     "closed",
	Firstpid:   "firstpid",
	Lastuid:    "lastuid",
	Lastpid:    "lastpid",
}

// NewBbsThreadDao creates and returns a new DAO object for table data access.
func NewBbsThreadDao(handlers ...gdb.ModelHandler) *BbsThreadDao {
	return &BbsThreadDao{
		group:    "default",
		table:    "bbs_thread",
		columns:  bbsThreadColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsThreadDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsThreadDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsThreadDao) Columns() BbsThreadColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsThreadDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsThreadDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsThreadDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
