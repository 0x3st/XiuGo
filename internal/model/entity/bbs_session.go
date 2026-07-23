// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsSession is the golang structure for table bbs_session.
type BbsSession struct {
	Sid       string `json:"sid"       orm:"sid"       description:""` //
	Uid       uint   `json:"uid"       orm:"uid"       description:""` //
	Fid       uint   `json:"fid"       orm:"fid"       description:""` //
	Url       string `json:"url"       orm:"url"       description:""` //
	Ip        uint   `json:"ip"        orm:"ip"        description:""` //
	Useragent string `json:"useragent" orm:"useragent" description:""` //
	Data      string `json:"data"      orm:"data"      description:""` //
	Bigdata   int    `json:"bigdata"   orm:"bigdata"   description:""` //
	LastDate  uint   `json:"lastDate"  orm:"last_date" description:""` //
}
