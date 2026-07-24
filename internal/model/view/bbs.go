package view

type ForumSummary struct {
	Fid           uint
	Name          string
	Brief         string
	Threads       uint
	Todayposts    uint
	Todaythreads  uint
	Announcement  string
	Icon          uint
	IconURL       string
}

type ThreadSummary struct {
	Tid           uint
	Uid           uint
	Fid           int
	Firstpid      uint
	ForumName     string
	Subject       string
	Username      string
	Avatar        uint
	AvatarURL     string
	CreateDate    uint
	CreateTime    string
	Lastuid       uint
	LastUsername  string
	LastDate      uint
	LastTime      string
	Views         uint
	Posts         uint
	Files         int
	Closed        uint
	Top           int
	TopClass      string
	AllowTop      bool
}

type PostView struct {
	Pid        uint
	Tid        uint
	Uid        uint
	Isfirst    uint
	Doctype    int
	Quotepid   int
	Floor      uint
	Username   string
	Avatar     uint
	AvatarURL  string
	CreateDate uint
	CreateTime string
	Message    string
	MessageFmt string
	CanQuote   bool
	CanEdit    bool
	CanDelete  bool
	Files      []Attachment
}

// Attachment is the public, post-associated shape used by the original
// Xiuno attachment list and download route.
type Attachment struct {
	Aid         uint
	Tid         int
	Pid         int
	Uid         int
	Filesize    uint
	Filename    string
	Orgfilename string
	Filetype    string
	Downloads   int
	Isimage     int
	URL         string
}

// PendingAttachment mirrors Xiuno's session-backed tmp_files entry. Path is
// kept server-side in the session and is never returned by the controller.
type PendingAttachment struct {
	ID          string `json:"aid"`
	URL         string `json:"url"`
	Path        string `json:"path"`
	Orgfilename string `json:"orgfilename"`
	Filetype    string `json:"filetype"`
	Filesize    uint   `json:"filesize"`
	Width       uint   `json:"width"`
	Height      uint   `json:"height"`
	Isimage     int    `json:"isimage"`
	Downloads   int    `json:"downloads"`
}

type ThreadPage struct {
	Thread      ThreadSummary
	First       PostView
	Replies     []PostView
	Author      ThreadAuthor
	CanReply    bool
	CanTop      bool
	ReplyNotice string
	NextFloor   uint
	Page        int
	Pages       int
	Pagination  ListPagination
	Keyword     string
}

type ThreadAuthor struct {
	Uid       uint
	Username  string
	Avatar    uint
	AvatarURL string
	Threads   int
	Posts     int
}

// ListPagination is passed to templates for original-style pager HTML.
type ListPagination struct {
	HTML     string
	Page     int
	Pages    int
	Total    int
	PageSize int
}

type Stats struct {
	Threads int
	Posts   int
	Users   int
	Onlines int
}

type User struct {
	Uid       uint
	Username  string
	Gid       uint
	Avatar    uint
	AvatarURL string
}

type AdminThread struct {
	Tid        uint
	Fid        int
	ForumName  string
	Subject    string
	Username   string
	CreateDate uint
	CreateTime string
	Views      uint
	Posts      uint
	Closed     uint
}

type AdminUser struct {
	Uid        uint
	Username   string
	Email      string
	Gid        uint
	GroupName  string
	Threads    int
	Posts      int
	Credits    int
	CreateDate uint
	CreateTime string
	CreateIp   uint
	CreateIP   string
}

type GroupOption struct {
	Gid  uint
	Name string
}

type PostEdit struct {
	Pid      uint
	Tid      uint
	Fid      int
	Subject  string
	Message  string
	Isfirst  uint
	Doctype  int
	Quotepid int
	Files    []Attachment
}

type UserProfile struct {
	Uid        uint
	Gid        uint
	Username   string
	Email      string
	GroupName  string
	Avatar     uint
	AvatarURL  string
	Threads    int
	Posts      int
	Credits    int
	Golds      int
	CreateDate uint
	CreateTime string
	LoginDate  uint
	LoginTime  string
}

type AdminForum struct {
	Fid          uint
	Name         string
	Rank         uint
	Threads      uint
	Todayposts   uint
	Brief        string
	Announcement string
	Accesson     uint
	Moduids      string
	Moderators   string
	Icon         uint
	IconURL      string
}

type ForumAccessRule struct {
	Gid         uint
	Name        string
	Allowread   uint
	Allowthread uint
	Allowpost   uint
	Allowattach uint
	Allowdown   uint
}

type GroupPermission struct {
	Gid             uint
	Name            string
	Creditsfrom     int
	Creditsto       int
	Allowread       int
	Allowthread     int
	Allowpost       int
	Allowattach     int
	Allowdown       int
	Allowtop        int
	Allowupdate     int
	Allowdelete     int
	Allowmove       int
	Allowbanuser    int
	Allowdeleteuser int
	Allowviewip     uint
}

// AdminDashboard mirrors the information blocks shown by Xiuno's original
// admin/index page. Field names deliberately follow the original concepts.
type AdminDashboard struct {
	Threads       int
	Posts         int
	Users         int
	Attachs       int
	Onlines       int
	DiskFreeSpace string
	OS            string
	WebServer     string
	GoVersion     string
	Database      string
	PostMaxSize   string
	UploadMaxSize string
	ClientIP      string
	ServerIP      string
	AllowURLFopen string
	SafeMode      string
	MaxExecution  string
	MemoryLimit   string
}

type SiteSettings struct {
	Sitename          string
	Sitebrief         string
	Runlevel          int
	RunlevelReason    string
	UserCreateOn      int
	UserCreateEmailOn int
	UserResetpwOn     int
	Lang              string
}

type SMTPAccount struct {
	Email string
	Host  string
	Port  int
	User  string
	Pass  string
}

type AdminUserPage struct {
	Users      []AdminUser
	SearchType string
	Keyword    string
	Page       int
	Pages      int
	Total      int
}

type AdminThreadScan struct {
	Fid             uint
	CreateDateStart uint
	CreateDateEnd   uint
	Uid             uint
	UserIP          uint
	Keyword         string
	Page            int
}

type AdminPlugin struct {
	Dir       string
	Name      string
	Brief     string
	Version   string
	Installed bool
	Enabled   bool
	HasConfig bool
}
