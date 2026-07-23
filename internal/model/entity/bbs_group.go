// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsGroup is the golang structure for table bbs_group.
type BbsGroup struct {
	Gid             uint   `json:"gid"             orm:"gid"             description:""` //
	Name            string `json:"name"            orm:"name"            description:""` //
	Creditsfrom     int    `json:"creditsfrom"     orm:"creditsfrom"     description:""` //
	Creditsto       int    `json:"creditsto"       orm:"creditsto"       description:""` //
	Allowread       int    `json:"allowread"       orm:"allowread"       description:""` //
	Allowthread     int    `json:"allowthread"     orm:"allowthread"     description:""` //
	Allowpost       int    `json:"allowpost"       orm:"allowpost"       description:""` //
	Allowattach     int    `json:"allowattach"     orm:"allowattach"     description:""` //
	Allowdown       int    `json:"allowdown"       orm:"allowdown"       description:""` //
	Allowtop        int    `json:"allowtop"        orm:"allowtop"        description:""` //
	Allowupdate     int    `json:"allowupdate"     orm:"allowupdate"     description:""` //
	Allowdelete     int    `json:"allowdelete"     orm:"allowdelete"     description:""` //
	Allowmove       int    `json:"allowmove"       orm:"allowmove"       description:""` //
	Allowbanuser    int    `json:"allowbanuser"    orm:"allowbanuser"    description:""` //
	Allowdeleteuser int    `json:"allowdeleteuser" orm:"allowdeleteuser" description:""` //
	Allowviewip     uint   `json:"allowviewip"     orm:"allowviewip"     description:""` //
}
