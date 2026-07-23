// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsForumAccessDao is the data access object for the table bbs_forum_access.
type BbsForumAccessDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  BbsForumAccessColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// BbsForumAccessColumns defines and stores column names for the table bbs_forum_access.
type BbsForumAccessColumns struct {
	Fid         string //
	Gid         string //
	Allowread   string //
	Allowthread string //
	Allowpost   string //
	Allowattach string //
	Allowdown   string //
}

// bbsForumAccessColumns holds the columns for the table bbs_forum_access.
var bbsForumAccessColumns = BbsForumAccessColumns{
	Fid:         "fid",
	Gid:         "gid",
	Allowread:   "allowread",
	Allowthread: "allowthread",
	Allowpost:   "allowpost",
	Allowattach: "allowattach",
	Allowdown:   "allowdown",
}

// NewBbsForumAccessDao creates and returns a new DAO object for table data access.
func NewBbsForumAccessDao(handlers ...gdb.ModelHandler) *BbsForumAccessDao {
	return &BbsForumAccessDao{
		group:    "default",
		table:    "bbs_forum_access",
		columns:  bbsForumAccessColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsForumAccessDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsForumAccessDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsForumAccessDao) Columns() BbsForumAccessColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsForumAccessDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsForumAccessDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsForumAccessDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
