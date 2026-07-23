// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BbsTableDay is the golang structure of table bbs_table_day for DAO operations like Where/Data.
type BbsTableDay struct {
	g.Meta     `orm:"table:bbs_table_day, do:true"`
	Year       any // 年
	Month      any // 月
	Day        any // 日
	CreateDate any // 时间戳
	Table      any // 表名
	Maxid      any // 最大ID
	Count      any // 总数
}
