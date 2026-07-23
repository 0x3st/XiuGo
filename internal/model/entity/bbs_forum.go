// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsForum is the golang structure for table bbs_forum.
type BbsForum struct {
	Fid          uint   `json:"fid"          orm:"fid"          description:""` //
	Name         string `json:"name"         orm:"name"         description:""` //
	Rank         uint   `json:"rank"         orm:"rank"         description:""` //
	Threads      uint   `json:"threads"      orm:"threads"      description:""` //
	Todayposts   uint   `json:"todayposts"   orm:"todayposts"   description:""` //
	Todaythreads uint   `json:"todaythreads" orm:"todaythreads" description:""` //
	Brief        string `json:"brief"        orm:"brief"        description:""` //
	Announcement string `json:"announcement" orm:"announcement" description:""` //
	Accesson     uint   `json:"accesson"     orm:"accesson"     description:""` //
	Orderby      int    `json:"orderby"      orm:"orderby"      description:""` //
	CreateDate   uint   `json:"createDate"   orm:"create_date"  description:""` //
	Icon         uint   `json:"icon"         orm:"icon"         description:""` //
	Moduids      string `json:"moduids"      orm:"moduids"      description:""` //
	SeoTitle     string `json:"seoTitle"     orm:"seo_title"    description:""` //
	SeoKeywords  string `json:"seoKeywords"  orm:"seo_keywords" description:""` //
}
