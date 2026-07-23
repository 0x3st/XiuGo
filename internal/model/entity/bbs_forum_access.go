// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsForumAccess is the golang structure for table bbs_forum_access.
type BbsForumAccess struct {
	Fid         uint `json:"fid"         orm:"fid"         description:""` //
	Gid         uint `json:"gid"         orm:"gid"         description:""` //
	Allowread   uint `json:"allowread"   orm:"allowread"   description:""` //
	Allowthread uint `json:"allowthread" orm:"allowthread" description:""` //
	Allowpost   uint `json:"allowpost"   orm:"allowpost"   description:""` //
	Allowattach uint `json:"allowattach" orm:"allowattach" description:""` //
	Allowdown   uint `json:"allowdown"   orm:"allowdown"   description:""` //
}
