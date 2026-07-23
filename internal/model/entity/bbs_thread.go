// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsThread is the golang structure for table bbs_thread.
type BbsThread struct {
	Fid        int    `json:"fid"        orm:"fid"         description:""` //
	Tid        uint   `json:"tid"        orm:"tid"         description:""` //
	Top        int    `json:"top"        orm:"top"         description:""` //
	Uid        uint   `json:"uid"        orm:"uid"         description:""` //
	Userip     uint   `json:"userip"     orm:"userip"      description:""` //
	Subject    string `json:"subject"    orm:"subject"     description:""` //
	CreateDate uint   `json:"createDate" orm:"create_date" description:""` //
	LastDate   uint   `json:"lastDate"   orm:"last_date"   description:""` //
	Views      uint   `json:"views"      orm:"views"       description:""` //
	Posts      uint   `json:"posts"      orm:"posts"       description:""` //
	Images     int    `json:"images"     orm:"images"      description:""` //
	Files      int    `json:"files"      orm:"files"       description:""` //
	Mods       int    `json:"mods"       orm:"mods"        description:""` //
	Closed     uint   `json:"closed"     orm:"closed"      description:""` //
	Firstpid   uint   `json:"firstpid"   orm:"firstpid"    description:""` //
	Lastuid    uint   `json:"lastuid"    orm:"lastuid"     description:""` //
	Lastpid    uint   `json:"lastpid"    orm:"lastpid"     description:""` //
}
