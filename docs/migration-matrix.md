# Xiuno BBS 4.0.4 → GoFrame 迁移矩阵

本清单以用户提供的 Xiuno BBS 4.0.4 PHP 源码为准。状态含义：

逐个 `$action`、不合并接口的审计口径见 [PHP 接口逐项台账](php-interface-ledger.md)。

- ✅ 已实现并验证
- 🟡 已实现基础流程，仍缺部分原版行为
- ⬜ 尚未实现

## 前台路由

| PHP 路由/动作 | 方法 | 原版职责 | 主要数据 | GoFrame 路由 | 状态 |
|---|---|---|---|---|---|
| `index` | GET | 首页、最新主题、运行统计 | forum/thread/post/user | `/` | ✅ |
| `forum-{fid}` | GET | 板块信息、主题列表、排序分页 | forum/thread | `/forum/:fid` | ✅ 访问权限、orderby、page 分页 |
| `thread-{tid}` | GET | 主题、首帖、回复、浏览数 | thread/post/user/attach | `/thread/:tid` | 🟡 首帖/回复分层、楼层、`message_fmt`、非图片附件列表与下载已对齐；分页与 return_html 片段已支持 |
| `thread-create-{fid}` | GET/POST | 发布主题 | thread/post/mythread/mypost/attach | `/thread/create` | ✅ 核心数据兼容；Session 临时附件随帖落库 |
| `post-create-{tid}` | GET/POST | 回复、引用、快速回复 | post/mypost/thread/forum/user/attach | `/thread/:tid/reply` | 🟡 回复、引用校验、doctype/`message_fmt`、附件随帖落库已实现；return_html 片段已支持 |
| `post-update-{pid}` | GET/POST | 编辑主题或回复、移动首帖 | post/thread/attach | `/post/:pid/edit` | 🟡 正文、标题、doctype 与追加/删除附件已实现；缺移动首帖 |
| `post-delete-{pid}` | POST | 删除回复或整篇主题 | post/thread/attach/index | `/post/:pid/delete` | ✅ 帖子、索引、计数与附件文件/记录级联清理已实现 |
| `attach-create` | POST | 上传附件、图片检测 | attach/upload/session | `POST /attach/create` | ✅ Base64 上传、类型白名单、20M 限制、Session `tmp_files` |
| `attach-delete-{aid}` | POST | 删除临时/正式附件 | attach/upload | `POST /attach/:aid/delete` | ✅ 临时 `_n` 与正式 aid；上传者/版主权限 |
| `attach-download-{aid}` | GET | 权限检查与附件下载 | attach/thread/forum | `GET /attach/:aid/download` | ✅ 板块 down 权限、`Content-Disposition`、downloads+1 |
| `user-{uid}` | GET | 用户资料 | user | `/user/:uid` | ✅ 基础资料 |
| `user-thread-{uid}` | GET | 用户主题 | user/thread | `/user/:uid/threads` | 🟡 缺分页与板块权限过滤 |
| `user-login` | GET/POST | Session/Token 登录 | user/session | `/login` | ✅ 浏览器预哈希、服务端拒收明文、Session 登录 |
| `user-create` | GET/POST | 注册、唯一性检查 | user | `/register` | 🟡 浏览器预哈希；无验证码注册及 PHP 登录已交叉验证，邮件分支待实投 |
| `user-logout` | GET | 注销 | session/token | `/logout` | ✅ |
| `user-resetpw*` | GET/POST | 邮件验证码与密码重置 | user/kv/mail | `/reset-password*` | 🟡 已实现，待真实 SMTP 验收 |
| `user-send_code-*` | POST | 注册/重置验证码 | kv/mail | `/register/send-code`、`/reset-password/send-code` | 🟡 已接原版 SMTP 配置，待实投 |
| `user-synlogin` | GET | 跨站同步登录 | session | — | ⬜ |
| `my` / `my-profile` | GET | 我的资料 | user | `/my` | ✅；原码中的 profile 分支实际被注释 |
| `my-password` | GET/POST | 修改密码 | user | `/my/password` | ✅ 浏览器预哈希，Go/PHP 双端密码兼容 |
| `my-thread` | GET | 我的主题 | mythread/thread | `/my/threads` | 🟡 缺分页 |
| `my-avatar` | GET/POST | 头像裁切上传 | user/upload | `/my/avatar` | ✅ base64 上传与默认头像 |
| `mod-top` | GET/POST | 置顶 | thread/thread_top/modlog | `POST /thread/:tid/top`、`POST /mod/top` | ✅ 0/1/3 范围、权限、bbs_thread_top、modlog、列表置顶排序 |
| `mod-close` | GET/POST | 关闭主题 | thread/modlog | `/admin/thread/:tid/close` | ✅ 管理端 |
| `mod-delete` | GET/POST | 批量删主题 | thread/post/index/modlog | `/admin/thread/:tid/delete` | ✅ 管理端基础版 |
| `mod-move` | GET/POST | 移动主题 | thread/forum/modlog | `POST /mod/move` | ✅ 批量移动 |
| `mod-deleteuser` | GET/POST | 删除用户及内容 | user/thread/post | — | ⬜ |
| `browser-download` | GET | 浏览器升级提示 | 静态文件 | — | ⬜ 非核心 |

## 管理后台路由

| PHP 路由/动作 | 原版职责 | GoFrame 路由 | 状态 |
|---|---|---|---|
| `admin/index` | 后台首页、环境与统计 | `/admin` | ✅ 基础统计 |
| `admin/index-login/logout/phpinfo` | 后台令牌、登出、环境 | 共用登录/`/logout` | 🟡 缺独立管理令牌、phpinfo |
| `admin/setting-base` | 站点名称、简介、运行级别等 | `/admin/?setting-base.htm` | 🟡 站名、简介、运行级别与用户开关已形成业务闭环；多语言未完 |
| `admin/setting-smtp` | SMTP 配置与测试 | `/admin/?setting-smtp.htm` | 🟡 保存与邮件发送已实现，待真实账号验收 |
| `admin/forum-list/update/getname/delete` | 板块 CRUD 与权限 | `/admin/?forum-*.htm`、`/admin/forum/*` | 🟡 兼容层已有批量新增/编辑/删除、图标、getname 与独立权限；待大批量验收 |
| `admin/group-list/update` | 用户组权限矩阵 | `/admin/?group-*.htm`、`/admin/group/:gid/edit` | 🟡 兼容层已有新增/删除/批量编辑和完整权限编辑；待批量验收 |
| `admin/user-list` | 搜索、分页、用户组展示 | `/admin/?user-list-*.htm`、`/admin/users` | 🟡 兼容层已有五类筛选与分页，待特殊条件验收 |
| `admin/user-create` | 创建用户 | `/admin/?user-create.htm` | 🟡 已实现，待 PHP 交叉验收 |
| `admin/user-update` | 用户名、邮箱、密码、用户组 | `/admin/?user-update-{uid}.htm`、`/admin/user/:uid/group` | 🟡 兼容层已覆盖全部核心字段，待交叉验收 |
| `admin/user-delete` | 删除非管理员用户 | `/admin/?user-delete.htm` | 🟡 已实现保护和内容/附件级联；超大账号仍待专项验收 |
| `admin/thread-list/scan/found` | 扫描、筛选、结果队列 | `/admin/?thread-*.htm`、`/admin/threads` | 🟡 已用 `bbs_queue` 实现完整筛选、分页扫描与结果页，待大数据量验收 |
| `admin/thread-operation` | 批量删除、关闭、打开 | `/admin/?thread-operation-*.htm`、现代单条路由 | 🟡 队列批量操作已实现；日志行为仍待逐项对齐 |
| `admin/other-cache` | 重建运行数据、清理上传临时目录 | `/admin/?other-cache.htm` | ✅ 单 Go：清 bbs_cache 后重算 runtime；清 upload/tmp（不再服务 PHP 并行） |
| `admin/plugin-*`（本地/市场） | Xiuno PHP 插件 | — | ❌ **不支持**；入口已移除 |
| （同上）官方插件市场 | — | — | ❌ 不迁移 |

## 核心数据兼容状态

| 表 | 用途 | 状态 |
|---|---|---|
| `bbs_forum` | 板块与计数 | ✅ 读取与发帖/回复计数 |
| `bbs_thread` | 主题索引 | ✅ 读、写、浏览数、关闭、删除 |
| `bbs_post` | 首帖和回复 | ✅ 读、写、随主题删除 |
| `bbs_user` | 用户、密码、计数、用户组 | ✅ 登录、计数、用户组调整 |
| `bbs_mythread` | 用户主题索引 | ✅ 新主题写入、删除清理 |
| `bbs_mypost` | 用户回复索引 | ✅ 新主题/回复写入、删除清理 |
| `bbs_modlog` | 管理操作日志 | ✅ Go 管理操作写入 |
| `bbs_group` | 用户组与权限 | ✅ 发帖、回复、版主权限及后台编辑 |
| `bbs_forum_access` | 板块独立权限覆盖 | ✅ Go/PHP 双端可见性验证 |
| `bbs_attach` | 附件元数据 | ✅ 上传落库、列表展示、下载计数、编辑删除与主题/用户级联清理 |
| `bbs_session*` | PHP Session | 🟡 Go 登录不复用；已按原版周期回收并重算在线数 |
| `bbs_cache` | 运行数据（runtime 等） | ✅ Go 读写 runtime；forumlist/grouplist 等 PHP 列表缓存已不再主动维护 |
| `bbs_queue` | 后台主题扫描队列 | 🟡 扫描、结果、批量操作和过期清理已实现；待大数据量验收 |
| `bbs_kv` | 验证码等短期数据 | 🟡 注册/找回验证码已使用；待真实 SMTP 验收 |
| `bbs_table_day` | 历史日统计 | ⬜ PHP 原版对应 cron 本身也被注释 |

## 原版计划任务兼容

- 每分钟检查一次原版到期条件，严格保留 `> 300` 秒和 `> 86400` 秒语义。
- 五分钟任务按 `online_hold_time` 清理 PHP Session、重算在线人数并更新 `cron_1_last_date`。
- 每日任务清零 runtime/板块今日计数，清理超过一天的临时附件和过期队列，并更新 `cron_2_last_date`。

## 建议迁移顺序

1. 主题、回复、登录、板块与用户索引完整兼容。
2. 管理后台主题/用户/板块管理与权限。
3. 注册、个人中心、密码修改和找回。
4. 附件上传下载、头像与图片处理。
5. 版主批量操作、缓存、队列和统计任务。
6. ~~PHP 插件~~ 已放弃；若需要扩展则设计 XiuGo 自有 Hook/模块 API。
