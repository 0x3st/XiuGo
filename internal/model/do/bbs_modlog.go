// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BbsModlog is the golang structure of table bbs_modlog for DAO operations like Where/Data.
type BbsModlog struct {
	g.Meta     `orm:"table:bbs_modlog, do:true"`
	Logid      any //
	Uid        any //
	Tid        any //
	Pid        any //
	Subject    any //
	Comment    any //
	Rmbs       any //
	CreateDate any //
	Action     any //
}
