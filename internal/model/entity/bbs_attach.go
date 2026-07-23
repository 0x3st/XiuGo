// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsAttach is the golang structure for table bbs_attach.
type BbsAttach struct {
	Aid         uint   `json:"aid"         orm:"aid"         description:""` //
	Tid         int    `json:"tid"         orm:"tid"         description:""` //
	Pid         int    `json:"pid"         orm:"pid"         description:""` //
	Uid         int    `json:"uid"         orm:"uid"         description:""` //
	Filesize    uint   `json:"filesize"    orm:"filesize"    description:""` //
	Width       uint   `json:"width"       orm:"width"       description:""` //
	Height      uint   `json:"height"      orm:"height"      description:""` //
	Filename    string `json:"filename"    orm:"filename"    description:""` //
	Orgfilename string `json:"orgfilename" orm:"orgfilename" description:""` //
	Filetype    string `json:"filetype"    orm:"filetype"    description:""` //
	CreateDate  uint   `json:"createDate"  orm:"create_date" description:""` //
	Comment     string `json:"comment"     orm:"comment"     description:""` //
	Downloads   int    `json:"downloads"   orm:"downloads"   description:""` //
	Credits     int    `json:"credits"     orm:"credits"     description:""` //
	Golds       int    `json:"golds"       orm:"golds"       description:""` //
	Rmbs        int    `json:"rmbs"        orm:"rmbs"        description:""` //
	Isimage     int    `json:"isimage"     orm:"isimage"     description:""` //
}
