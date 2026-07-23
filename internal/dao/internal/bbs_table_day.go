// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsTableDayDao is the data access object for the table bbs_table_day.
type BbsTableDayDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsTableDayColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsTableDayColumns defines and stores column names for the table bbs_table_day.
type BbsTableDayColumns struct {
	Year       string // 年
	Month      string // 月
	Day        string // 日
	CreateDate string // 时间戳
	Table      string // 表名
	Maxid      string // 最大ID
	Count      string // 总数
}

// bbsTableDayColumns holds the columns for the table bbs_table_day.
var bbsTableDayColumns = BbsTableDayColumns{
	Year:       "year",
	Month:      "month",
	Day:        "day",
	CreateDate: "create_date",
	Table:      "table",
	Maxid:      "maxid",
	Count:      "count",
}

// NewBbsTableDayDao creates and returns a new DAO object for table data access.
func NewBbsTableDayDao(handlers ...gdb.ModelHandler) *BbsTableDayDao {
	return &BbsTableDayDao{
		group:    "default",
		table:    "bbs_table_day",
		columns:  bbsTableDayColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsTableDayDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsTableDayDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsTableDayDao) Columns() BbsTableDayColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsTableDayDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsTableDayDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsTableDayDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
