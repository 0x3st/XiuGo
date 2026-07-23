// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BbsTableDay is the golang structure for table bbs_table_day.
type BbsTableDay struct {
	Year       uint   `json:"year"       orm:"year"        description:"年"`    // 年
	Month      uint   `json:"month"      orm:"month"       description:"月"`    // 月
	Day        uint   `json:"day"        orm:"day"         description:"日"`    // 日
	CreateDate uint   `json:"createDate" orm:"create_date" description:"时间戳"`  // 时间戳
	Table      string `json:"table"      orm:"table"       description:"表名"`   // 表名
	Maxid      uint   `json:"maxid"      orm:"maxid"       description:"最大ID"` // 最大ID
	Count      uint   `json:"count"      orm:"count"       description:"总数"`   // 总数
}
