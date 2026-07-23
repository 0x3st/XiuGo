// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsForumDao is the data access object for the table bbs_forum.
type BbsForumDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsForumColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsForumColumns defines and stores column names for the table bbs_forum.
type BbsForumColumns struct {
	Fid          string //
	Name         string //
	Rank         string //
	Threads      string //
	Todayposts   string //
	Todaythreads string //
	Brief        string //
	Announcement string //
	Accesson     string //
	Orderby      string //
	CreateDate   string //
	Icon         string //
	Moduids      string //
	SeoTitle     string //
	SeoKeywords  string //
}

// bbsForumColumns holds the columns for the table bbs_forum.
var bbsForumColumns = BbsForumColumns{
	Fid:          "fid",
	Name:         "name",
	Rank:         "rank",
	Threads:      "threads",
	Todayposts:   "todayposts",
	Todaythreads: "todaythreads",
	Brief:        "brief",
	Announcement: "announcement",
	Accesson:     "accesson",
	Orderby:      "orderby",
	CreateDate:   "create_date",
	Icon:         "icon",
	Moduids:      "moduids",
	SeoTitle:     "seo_title",
	SeoKeywords:  "seo_keywords",
}

// NewBbsForumDao creates and returns a new DAO object for table data access.
func NewBbsForumDao(handlers ...gdb.ModelHandler) *BbsForumDao {
	return &BbsForumDao{
		group:    "default",
		table:    "bbs_forum",
		columns:  bbsForumColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsForumDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsForumDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsForumDao) Columns() BbsForumColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsForumDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsForumDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsForumDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
