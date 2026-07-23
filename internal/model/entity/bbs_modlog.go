// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsModlog is the golang structure for table bbs_modlog.
type BbsModlog struct {
	Logid      uint   `json:"logid"      orm:"logid"       description:""` //
	Uid        uint   `json:"uid"        orm:"uid"         description:""` //
	Tid        uint   `json:"tid"        orm:"tid"         description:""` //
	Pid        uint   `json:"pid"        orm:"pid"         description:""` //
	Subject    string `json:"subject"    orm:"subject"     description:""` //
	Comment    string `json:"comment"    orm:"comment"     description:""` //
	Rmbs       int    `json:"rmbs"       orm:"rmbs"        description:""` //
	CreateDate uint   `json:"createDate" orm:"create_date" description:""` //
	Action     string `json:"action"     orm:"action"      description:""` //
}
