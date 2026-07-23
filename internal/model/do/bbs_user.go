// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BbsUser is the golang structure of table bbs_user for DAO operations like Where/Data.
type BbsUser struct {
	g.Meta      `orm:"table:bbs_user, do:true"`
	Uid         any // 用户编号
	Gid         any // 用户组编号
	Email       any // 邮箱
	Username    any // 用户名
	Realname    any // 用户名
	Idnumber    any // 用户名
	Password    any // 密码
	PasswordSms any // 密码
	Salt        any // 密码混杂
	Mobile      any // 手机号
	Qq          any // QQ
	Threads     any // 发帖数
	Posts       any // 回帖数
	Credits     any // 积分
	Golds       any // 金币
	Rmbs        any // 人民币
	CreateIp    any // 创建时IP
	CreateDate  any // 创建时间
	LoginIp     any // 登录时IP
	LoginDate   any // 登录时间
	Logins      any // 登录次数
	Avatar      any // 用户最后更新图像时间
}
