// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsUser is the golang structure for table bbs_user.
type BbsUser struct {
	Uid         uint   `json:"uid"         orm:"uid"          description:"用户编号"`       // 用户编号
	Gid         uint   `json:"gid"         orm:"gid"          description:"用户组编号"`      // 用户组编号
	Email       string `json:"email"       orm:"email"        description:"邮箱"`         // 邮箱
	Username    string `json:"username"    orm:"username"     description:"用户名"`        // 用户名
	Realname    string `json:"realname"    orm:"realname"     description:"用户名"`        // 用户名
	Idnumber    string `json:"idnumber"    orm:"idnumber"     description:"用户名"`        // 用户名
	Password    string `json:"password"    orm:"password"     description:"密码"`         // 密码
	PasswordSms string `json:"passwordSms" orm:"password_sms" description:"密码"`         // 密码
	Salt        string `json:"salt"        orm:"salt"         description:"密码混杂"`       // 密码混杂
	Mobile      string `json:"mobile"      orm:"mobile"       description:"手机号"`        // 手机号
	Qq          string `json:"qq"          orm:"qq"           description:"QQ"`         // QQ
	Threads     int    `json:"threads"     orm:"threads"      description:"发帖数"`        // 发帖数
	Posts       int    `json:"posts"       orm:"posts"        description:"回帖数"`        // 回帖数
	Credits     int    `json:"credits"     orm:"credits"      description:"积分"`         // 积分
	Golds       int    `json:"golds"       orm:"golds"        description:"金币"`         // 金币
	Rmbs        int    `json:"rmbs"        orm:"rmbs"         description:"人民币"`        // 人民币
	CreateIp    uint   `json:"createIp"    orm:"create_ip"    description:"创建时IP"`      // 创建时IP
	CreateDate  uint   `json:"createDate"  orm:"create_date"  description:"创建时间"`       // 创建时间
	LoginIp     uint   `json:"loginIp"     orm:"login_ip"     description:"登录时IP"`      // 登录时IP
	LoginDate   uint   `json:"loginDate"   orm:"login_date"   description:"登录时间"`       // 登录时间
	Logins      uint   `json:"logins"      orm:"logins"       description:"登录次数"`       // 登录次数
	Avatar      uint   `json:"avatar"      orm:"avatar"       description:"用户最后更新图像时间"` // 用户最后更新图像时间
}
