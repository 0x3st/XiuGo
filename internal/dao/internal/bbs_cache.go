// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsCacheDao is the data access object for the table bbs_cache.
type BbsCacheDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsCacheColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsCacheColumns defines and stores column names for the table bbs_cache.
type BbsCacheColumns struct {
	K      string //
	V      string //
	Expiry string //
}

// bbsCacheColumns holds the columns for the table bbs_cache.
var bbsCacheColumns = BbsCacheColumns{
	K:      "k",
	V:      "v",
	Expiry: "expiry",
}

// NewBbsCacheDao creates and returns a new DAO object for table data access.
func NewBbsCacheDao(handlers ...gdb.ModelHandler) *BbsCacheDao {
	return &BbsCacheDao{
		group:    "default",
		table:    "bbs_cache",
		columns:  bbsCacheColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsCacheDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsCacheDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsCacheDao) Columns() BbsCacheColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsCacheDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsCacheDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsCacheDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
