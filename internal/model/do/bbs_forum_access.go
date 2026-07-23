// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BbsForumAccess is the golang structure of table bbs_forum_access for DAO operations like Where/Data.
type BbsForumAccess struct {
	g.Meta      `orm:"table:bbs_forum_access, do:true"`
	Fid         any //
	Gid         any //
	Allowread   any //
	Allowthread any //
	Allowpost   any //
	Allowattach any //
	Allowdown   any //
}
