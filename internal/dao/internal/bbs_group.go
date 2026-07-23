// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsGroupDao is the data access object for the table bbs_group.
type BbsGroupDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsGroupColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsGroupColumns defines and stores column names for the table bbs_group.
type BbsGroupColumns struct {
	Gid             string //
	Name            string //
	Creditsfrom     string //
	Creditsto       string //
	Allowread       string //
	Allowthread     string //
	Allowpost       string //
	Allowattach     string //
	Allowdown       string //
	Allowtop        string //
	Allowupdate     string //
	Allowdelete     string //
	Allowmove       string //
	Allowbanuser    string //
	Allowdeleteuser string //
	Allowviewip     string //
}

// bbsGroupColumns holds the columns for the table bbs_group.
var bbsGroupColumns = BbsGroupColumns{
	Gid:             "gid",
	Name:            "name",
	Creditsfrom:     "creditsfrom",
	Creditsto:       "creditsto",
	Allowread:       "allowread",
	Allowthread:     "allowthread",
	Allowpost:       "allowpost",
	Allowattach:     "allowattach",
	Allowdown:       "allowdown",
	Allowtop:        "allowtop",
	Allowupdate:     "allowupdate",
	Allowdelete:     "allowdelete",
	Allowmove:       "allowmove",
	Allowbanuser:    "allowbanuser",
	Allowdeleteuser: "allowdeleteuser",
	Allowviewip:     "allowviewip",
}

// NewBbsGroupDao creates and returns a new DAO object for table data access.
func NewBbsGroupDao(handlers ...gdb.ModelHandler) *BbsGroupDao {
	return &BbsGroupDao{
		group:    "default",
		table:    "bbs_group",
		columns:  bbsGroupColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsGroupDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsGroupDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsGroupDao) Columns() BbsGroupColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsGroupDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsGroupDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsGroupDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
