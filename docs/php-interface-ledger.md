# Xiuno BBS 4.0.4 PHP 接口逐项台账

本台账以 `/Users/laywoo/xp/xiuno-bbs` 中用户提供的源码为唯一依据，按
`route/*.php`、`admin/route/*.php` 的实际 `$action` 分支逐项登记。它与按功能归类的
`migration-matrix.md` 不同：这里不使用 `resetpw*`、`plugin-*` 之类的合并项。

状态：✅ 核心行为已实现并交叉验证；🟡 已有对应入口但仍缺原版行为；⬜ 尚无 Go 对应入口。

## 前台：30 项

| # | PHP 文件与 action | 方法 | 关键输入/行为 | GoFrame 对应 | 状态 |
|---:|---|---|---|---|---|
| 1 | `route/index.php` 默认 | GET | 首页、最新主题、统计 | `GET /` | ✅ |
| 2 | `route/forum.php` 默认 | GET | `fid/page/order`、板块主题列表 | `GET /forum/:fid` | 🟡 已有板块访问权限；缺分页、排序 |
| 3 | `route/thread.php` 默认 | GET | `tid/page/keyword`、主题和帖子 | `GET /thread/:tid` | 🟡 首帖/回复、楼层、`message_fmt`、引用入口和权限已对齐；缺分页、关键词定位、附件 |
| 4 | `route/thread.php:create` | GET/POST | `fid/subject/message/doctype` | `ALL /thread/create` | ✅ 核心表、用户索引、计数与 Session 临时附件落库兼容 |
| 5 | `route/post.php:create` | GET/POST | `tid/quick/quotepid/message/doctype/return_html` | `POST /thread/:tid/reply` | 🟡 引用归属校验、引用摘要、doctype、`message_fmt` 与附件落库已实现；缺 HTML 片段响应 |
| 6 | `route/post.php:update` | GET/POST | `pid/fid/subject/message/doctype` | `ALL /post/:pid/edit` | 🟡 已支持正文、首帖标题、doctype 与追加/删除附件；缺移动板块 |
| 7 | `route/post.php:delete` | POST | `pid`；首帖删主题，回复单删 | `POST /post/:pid/delete` | ✅ 主题/回复、索引计数与附件文件/记录级联清理 |
| 8 | `route/attach.php:create` | POST | `width/height/is_image/name/data` | `POST /attach/create` | ✅ Base64 上传、类型白名单、20M、Session 临时区 |
| 9 | `route/attach.php:delete` | POST | `aid`、上传者/版主权限 | `POST /attach/:aid/delete` | ✅ 临时 `_n` 与正式 aid；上传者/版主 |
| 10 | `route/attach.php:download` | GET | `aid`、板块下载权限、文件响应 | `GET /attach/:aid/download` | ✅ down 权限、强制下载头、downloads+1 |
| 11 | `route/browser.php:download` | GET | `type`、浏览器下载跳转 | — | ⬜ 非核心 |
| 12 | `route/mod.php:top` | GET/POST | `top/tidarr[]`、置顶与日志 | `POST /thread/:tid/top`、`POST /mod/top` | ✅ 0 取消/1 版块/3 全站；admin 全站限制；modlog；列表前置 |
| 13 | `route/mod.php:close` | GET/POST | `close/tidarr[]` | `POST /admin/thread/:tid/close` | 🟡 管理端支持单条关闭/打开 |
| 14 | `route/mod.php:delete` | GET/POST | `tidarr[]`、批量删除 | `POST /admin/thread/:tid/delete` | 🟡 管理端支持单条删除 |
| 15 | `route/mod.php:move` | GET/POST | `tidarr[]/newfid` | — | ⬜ |
| 16 | `route/mod.php:deleteuser` | POST | `uid`、删用户及全部内容 | — | ⬜ |
| 17 | `route/my.php` 默认 | GET | 当前用户资料首页 | `GET /my` | ✅ |
| 18 | `route/my.php:password` | GET/POST | `password_old/password_new/password_new_repeat` | `ALL /my/password` | ✅ 浏览器预哈希、服务端拒收明文，双端密码格式兼容 |
| 19 | `route/my.php:thread` | GET | `page`、`bbs_mythread` 主题列表 | `GET /my/threads` | 🟡 缺分页 |
| 20 | `route/my.php:avatar` | GET/POST | `width/height/data`、头像文件 | — | ⬜ |
| 21 | `route/user.php` 默认 | GET | `uid`、公开资料 | `GET /user/:uid` | ✅ 基础资料 |
| 22 | `route/user.php:thread` | GET | `uid/page`、公开主题 | `GET /user/:uid/threads` | 🟡 缺分页和板块访问过滤 |
| 23 | `route/user.php:login` | GET/POST | 浏览器先 MD5；PHP 再 `md5(clientHash+salt)` | `ALL /login` | 🟡 原版客户端哈希协议已强制执行并实测；未复用 PHP token |
| 24 | `route/user.php:create` | GET/POST | 邮箱、用户名、密码、可选验证码 | `ALL /register` | 🟡 浏览器只提交客户端哈希；无验证码注册与 PHP 登录已交叉验证，SMTP 分支待实投 |
| 25 | `route/user.php:logout` | GET | Session 与 token 清理 | `POST /logout` | 🟡 Go Session 已实现，PHP token 不复用 |
| 26 | `route/user.php:resetpw` | GET/POST | 邮箱与验证码校验 | `ALL /reset-password` | 🟡 5 分钟 Session 校验已实现，待真实 SMTP 验收 |
| 27 | `route/user.php:resetpw_complete` | GET/POST | 验证通过后写新密码 | `ALL /reset-password/complete` | 🟡 浏览器只提交客户端哈希，原 salt 与密码格式兼容；待邮件全链路验收 |
| 28 | `route/user.php:send_code/user_create` | POST | 注册邮箱验证码 | `POST /register/send-code` | 🟡 已接原版 SMTP 配置，当前无账号可实投 |
| 29 | `route/user.php:send_code/user_resetpw` | POST | 找回密码邮箱验证码 | `POST /reset-password/send-code` | 🟡 已接原版 SMTP 配置，当前无账号可实投 |
| 30 | `route/user.php:synlogin` | GET | 加密 token 与回跳地址 | — | ⬜ |

> `route/my.php:profile` 在该源码中被整段注释，不计为有效接口。

## 后台：35 项

| # | PHP 文件与 action | 方法 | 关键输入/行为 | GoFrame 对应 | 状态 |
|---:|---|---|---|---|---|
| 31 | `admin/route/index.php` 默认 | GET | 环境信息、主题/帖子/用户/附件统计 | `GET /admin` | 🟡 已有核心统计，缺环境和版本检查 |
| 32 | `admin/route/index.php:login` | GET/POST | 二次管理员密码与 admin token | `/admin/?index-login.htm` | 🟡 客户端哈希与二次验证已实现；无独立 PHP admin token |
| 33 | `admin/route/index.php:logout` | GET | 清理 admin token | `POST /logout` | 🟡 共用 Session 退出 |
| 34 | `admin/route/index.php:phpinfo` | GET | PHP 环境输出 | — | ⬜ Go 版不应照搬 phpinfo |
| 35 | `admin/route/setting.php:base` | GET/POST | 站名、简介、运行级别、注册开关、语言 | `/admin/?setting-base.htm` | 🟡 站名、简介、运行级别和三个用户开关已生效；多语言页面仍未完整切换 |
| 36 | `admin/route/setting.php:smtp` | GET/POST | SMTP 账号数组与测试 | `/admin/?setting-smtp.htm` | 🟡 保存和发送实现已接入；缺后台“测试 SMTP”按钮的实投验收 |
| 37 | `admin/route/forum.php:list` | GET/POST | 板块列表、批量名称/排序/图标更新 | `/admin/?forum-list.htm`、`GET /admin/forums` | 🟡 兼容层已实现批量新增/更新、图标和删除；待大批量与失败回滚验收 |
| 38 | `admin/route/forum.php:update` | GET/POST | 板块资料、版主、访问权限矩阵 | `ALL /admin/forum/:fid/edit` | ✅ 板块资料、版主、权限和 PHP 缓存兼容 |
| 39 | `admin/route/forum.php:getname` | GET | `uids` 转用户名 | `/admin/?forum-getname-{uids}.htm` | 🟡 已实现 UID 到版主用户名解析，待浏览器交叉验收 |
| 40 | `admin/route/forum.php:delete` | GET/POST | `fid`、删除板块及主题 | `/admin/?forum-delete-{fid}.htm`、`POST /admin/forum/:fid/delete` | 🟡 兼容层支持有主题板块级联删除；现代入口仍只允许空板块 |
| 41 | `admin/route/group.php:list` | GET/POST | 用户组列表、名称和积分范围 | `/admin/?group-list.htm`、`GET /admin/groups` | 🟡 兼容层已实现批量新增/更新/删除并维护板块权限；待大批量验收 |
| 42 | `admin/route/group.php:update` | GET/POST | 阅读/发帖/附件/版主权限矩阵 | `ALL /admin/group/:gid/edit` | ✅ 现有用户组完整权限编辑和 PHP 缓存兼容 |
| 43 | `admin/route/user.php:list` | GET | `srchtype/keyword/page` | `/admin/?user-list-{type}-{keyword}-{page}.htm`、`GET /admin/users` | 🟡 兼容层支持 UID/用户名/邮箱/组/IP 筛选与分页；待特殊字符验收 |
| 44 | `admin/route/user.php:create` | GET/POST | 邮箱、用户名、密码、用户组 | `/admin/?user-create.htm` | 🟡 已实现客户端预哈希、原版存储格式、唯一性和 runtime 计数，待 PHP 端交叉验收 |
| 45 | `admin/route/user.php:update` | GET/POST | 邮箱、用户名、可选新密码、用户组 | `/admin/?user-update-{uid}.htm`、`POST /admin/user/:uid/group` | 🟡 兼容层支持邮箱、用户名、客户端预哈希密码和组；待全字段交叉验收 |
| 46 | `admin/route/user.php:delete` | POST | `uid`、禁止删除管理员、级联内容 | `/admin/?user-delete.htm` | 🟡 已实现管理员保护及主题/回复/附件级联；超大账号与计数副作用仍需专项验收 |
| 47 | `admin/route/thread.php:list` | GET | 初始化扫描队列与板块摘要 | `/admin/?thread-list.htm`、`GET /admin/threads` | 🟡 兼容层已初始化 `bbs_queue` 扫描会话 |
| 48 | `admin/route/thread.php:scan` | POST/GET | `fid/date/uid/userip/keyword/page` 扫描 | `/admin/?thread-scan.htm` | 🟡 已实现原版筛选项、分页扫描并写队列；待大数据量验收 |
| 49 | `admin/route/thread.php:operation/delete` | POST/GET | 从队列批量删除 | `/admin/?thread-operation-delete.htm`、`POST /admin/thread/:tid/delete` | 🟡 已实现队列批量删除和完整回复计数回补；操作日志仍待逐项对齐 |
| 50 | `admin/route/thread.php:operation/close` | POST/GET | 从队列批量关闭 | `/admin/?thread-operation-close.htm`、`POST /admin/thread/:tid/close` | 🟡 已实现队列批量关闭；操作日志仍待逐项对齐 |
| 51 | `admin/route/thread.php:operation/open` | POST/GET | 从队列批量打开 | `/admin/?thread-operation-open.htm`、`POST /admin/thread/:tid/close` | 🟡 已实现队列批量打开；操作日志仍待逐项对齐 |
| 52 | `admin/route/thread.php:found` | GET | 队列结果分页 | `/admin/?thread-found-{page}.htm` | 🟡 已实现队列结果和分页，待浏览器交叉验收 |
| 53 | `admin/route/other.php:cache` | GET/POST | 重建 runtime、清 upload/tmp | `/admin/?other-cache.htm` | ✅ 单 Go 维护工具；不再以 PHP 编译缓存为准 |
| 54 | `admin/route/plugin.php:local` | GET | 本地插件列表 | `/admin/?plugin-local.htm` | 🟡 已读取 `plugin/*/conf.json` 并展示本地状态；不执行 PHP Hook |
| 55 | `admin/route/plugin.php:official_fee` | GET | 官方收费插件分页 | — | ⬜ |
| 56 | `admin/route/plugin.php:official_free` | GET | 官方免费插件分页 | — | ⬜ |
| 57 | `admin/route/plugin.php:read` | GET | 插件详情和购买状态 | — | ⬜ |
| 58 | `admin/route/plugin.php:is_bought` | GET | 购买状态查询 | — | ⬜ |
| 59 | `admin/route/plugin.php:download` | GET/POST | 下载并解压官方插件 | — | ⬜ |
| 60 | `admin/route/plugin.php:install` | GET/POST | 依赖检查、安装脚本、同类互斥 | `/admin/?plugin-install-{dir}.htm` | 🟡 本地配置状态、依赖检查和同类互斥已实现；不执行 `install.php` |
| 61 | `admin/route/plugin.php:unstall` | GET/POST | 依赖检查、卸载脚本 | `/admin/?plugin-unstall-{dir}.htm` | 🟡 本地配置状态和反向依赖检查已实现；不执行 `unstall.php` |
| 62 | `admin/route/plugin.php:enable` | GET/POST | 启用插件 | `/admin/?plugin-enable-{dir}.htm` | 🟡 可更新本地启用状态；Go 尚无 PHP Hook/Overwrite 执行层 |
| 63 | `admin/route/plugin.php:disable` | GET/POST | 禁用插件 | `/admin/?plugin-disable-{dir}.htm` | 🟡 可更新本地禁用状态和检查依赖；Go 尚无 PHP Hook/Overwrite 执行层 |
| 64 | `admin/route/plugin.php:upgrade` | GET/POST | 下载、升级与升级脚本 | — | ⬜ |
| 65 | `admin/route/plugin.php:setting` | GET/POST | 执行插件自带设置入口 | — | ⬜ |

## 统计口径

- 严格登记的有效接口/action 变体：65 项（前台 30、后台 35）。
- Go 路由数量不是完成度：一个 Go 路由可能对应 PHP 的 GET/POST 两种行为，也可能只覆盖原接口的一部分。
- 只有经过数据库索引/计数核对，并由 PHP 原版页面或接口交叉读取成功的项目，才可标为 ✅。
