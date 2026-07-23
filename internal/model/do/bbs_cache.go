// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BbsCache is the golang structure of table bbs_cache for DAO operations like Where/Data.
type BbsCache struct {
	g.Meta `orm:"table:bbs_cache, do:true"`
	K      any //
	V      any //
	Expiry any //
}
