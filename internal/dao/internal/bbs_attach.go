// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsAttachDao is the data access object for the table bbs_attach.
type BbsAttachDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsAttachColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsAttachColumns defines and stores column names for the table bbs_attach.
type BbsAttachColumns struct {
	Aid         string //
	Tid         string //
	Pid         string //
	Uid         string //
	Filesize    string //
	Width       string //
	Height      string //
	Filename    string //
	Orgfilename string //
	Filetype    string //
	CreateDate  string //
	Comment     string //
	Downloads   string //
	Credits     string //
	Golds       string //
	Rmbs        string //
	Isimage     string //
}

// bbsAttachColumns holds the columns for the table bbs_attach.
var bbsAttachColumns = BbsAttachColumns{
	Aid:         "aid",
	Tid:         "tid",
	Pid:         "pid",
	Uid:         "uid",
	Filesize:    "filesize",
	Width:       "width",
	Height:      "height",
	Filename:    "filename",
	Orgfilename: "orgfilename",
	Filetype:    "filetype",
	CreateDate:  "create_date",
	Comment:     "comment",
	Downloads:   "downloads",
	Credits:     "credits",
	Golds:       "golds",
	Rmbs:        "rmbs",
	Isimage:     "isimage",
}

// NewBbsAttachDao creates and returns a new DAO object for table data access.
func NewBbsAttachDao(handlers ...gdb.ModelHandler) *BbsAttachDao {
	return &BbsAttachDao{
		group:    "default",
		table:    "bbs_attach",
		columns:  bbsAttachColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsAttachDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsAttachDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsAttachDao) Columns() BbsAttachColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsAttachDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsAttachDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsAttachDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
