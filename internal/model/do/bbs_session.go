// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BbsSession is the golang structure of table bbs_session for DAO operations like Where/Data.
type BbsSession struct {
	g.Meta    `orm:"table:bbs_session, do:true"`
	Sid       any //
	Uid       any //
	Fid       any //
	Url       any //
	Ip        any //
	Useragent any //
	Data      any //
	Bigdata   any //
	LastDate  any //
}
