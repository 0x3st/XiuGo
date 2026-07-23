// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BbsUserDao is the data access object for the table bbs_user.
type BbsUserDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BbsUserColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BbsUserColumns defines and stores column names for the table bbs_user.
type BbsUserColumns struct {
	Uid         string // 用户编号
	Gid         string // 用户组编号
	Email       string // 邮箱
	Username    string // 用户名
	Realname    string // 用户名
	Idnumber    string // 用户名
	Password    string // 密码
	PasswordSms string // 密码
	Salt        string // 密码混杂
	Mobile      string // 手机号
	Qq          string // QQ
	Threads     string // 发帖数
	Posts       string // 回帖数
	Credits     string // 积分
	Golds       string // 金币
	Rmbs        string // 人民币
	CreateIp    string // 创建时IP
	CreateDate  string // 创建时间
	LoginIp     string // 登录时IP
	LoginDate   string // 登录时间
	Logins      string // 登录次数
	Avatar      string // 用户最后更新图像时间
}

// bbsUserColumns holds the columns for the table bbs_user.
var bbsUserColumns = BbsUserColumns{
	Uid:         "uid",
	Gid:         "gid",
	Email:       "email",
	Username:    "username",
	Realname:    "realname",
	Idnumber:    "idnumber",
	Password:    "password",
	PasswordSms: "password_sms",
	Salt:        "salt",
	Mobile:      "mobile",
	Qq:          "qq",
	Threads:     "threads",
	Posts:       "posts",
	Credits:     "credits",
	Golds:       "golds",
	Rmbs:        "rmbs",
	CreateIp:    "create_ip",
	CreateDate:  "create_date",
	LoginIp:     "login_ip",
	LoginDate:   "login_date",
	Logins:      "logins",
	Avatar:      "avatar",
}

// NewBbsUserDao creates and returns a new DAO object for table data access.
func NewBbsUserDao(handlers ...gdb.ModelHandler) *BbsUserDao {
	return &BbsUserDao{
		group:    "default",
		table:    "bbs_user",
		columns:  bbsUserColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BbsUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BbsUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BbsUserDao) Columns() BbsUserColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BbsUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BbsUserDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BbsUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
