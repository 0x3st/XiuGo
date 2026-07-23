// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsPost is the golang structure for table bbs_post.
type BbsPost struct {
	Tid        uint   `json:"tid"        orm:"tid"         description:""` //
	Pid        uint   `json:"pid"        orm:"pid"         description:""` //
	Uid        uint   `json:"uid"        orm:"uid"         description:""` //
	Isfirst    uint   `json:"isfirst"    orm:"isfirst"     description:""` //
	CreateDate uint   `json:"createDate" orm:"create_date" description:""` //
	Userip     uint   `json:"userip"     orm:"userip"      description:""` //
	Images     int    `json:"images"     orm:"images"      description:""` //
	Files      int    `json:"files"      orm:"files"       description:""` //
	Doctype    int    `json:"doctype"    orm:"doctype"     description:""` //
	Quotepid   int    `json:"quotepid"   orm:"quotepid"    description:""` //
	Message    string `json:"message"    orm:"message"     description:""` //
	MessageFmt string `json:"messageFmt" orm:"message_fmt" description:""` //
}
