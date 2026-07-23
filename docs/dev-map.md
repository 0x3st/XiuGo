# XiuGo 开发地图（模板 / Theme / Plugin）

面向「没有页面上 hook 标记会很煎熬」的开发体验：先查表，再改文件。

## 1. 改某个 URL 长什么样？

前台走 `renderPage`（会应用 **Theme CSS**、**Theme 模板覆盖**、**page.render** 插件）：

| URL | 模板 |
|-----|------|
| `GET /` | `pages/home.html` |
| `GET /forum/:fid` | `pages/forum.html` |
| `GET /thread/:tid` | `pages/thread.html` |
| `GET/POST /thread/create` | `pages/create.html` |
| `GET /thread/:tid/reply` | `pages/create.html`（高级回帖） |
| `ALL /post/:pid/edit` | `pages/post_edit.html` |
| `ALL /login` | `pages/login.html` |
| `ALL /register` | `pages/register.html` |
| `ALL /reset-password*` | `pages/reset_password*.html` |
| `GET /user/:uid`、`GET /my` | `pages/user_profile.html` |
| `GET /user/:uid/threads`、`GET /my/threads` | `pages/user_threads.html` |
| `ALL /my/password` | `pages/my_password.html` |
| `ALL /my/avatar` | `pages/my_avatar.html` |

共用：

| 文件 | 作用 |
|------|------|
| `layout/header.html` | HTML 头、引入 CSS、ThemeCSS/PluginCSS |
| `layout/header_nav.html` | 顶栏 |
| `layout/footer.html` | 脚本、PluginFooterHTML |
| `partials/thread_list.html` | 主题列表行 |
| `partials/post_list_item.html` | 回帖行 / ajax 片段 |
| `partials/thread_list_mod.html` | 列表版主条 |

后台多数直接 `WriteTpl`（**暂不走** theme 覆盖 / page.render）：

- `admin/*.html`、`admin_compat/*.html`、`mod/*.html`

在仓库里定位：

```bash
grep -rn 'pages/thread.html' internal/controller
```

## 2. 做 Theme（皮肤）

```text
resource/themes/<id>/
  theme.json                 # name, version, css: ["css/theme.css"]
  css/theme.css              # 覆盖默认 Bootstrap-BBS
  templates/                 # 可选：覆盖 HTML
    pages/home.html
    README.txt
```

- 静态访问：`/themes/<id>/css/theme.css`
- 激活：后台 → **主题**，存 `bbs_kv.xiugo_theme`
- **模板覆盖**：若存在 `themes/<id>/templates/<与 resource/template 相同相对路径>`，则用主题文件；否则用默认
- Theme **不靠** hook；主要是 CSS + 可选整页 HTML

## 3. 做 Plugin（Go 扩展）

```text
internal/plugin/plugin.go
internal/plugin/builtin/hello.go      # 最小示例
internal/plugin/builtin/devtools.go   # 页脚注释显示当前模板路径
```

1. 实现 `plugin.Plugin`，`init()` 里 `Register` + `On`
2. 保证被 import（`cmd` 已 `_ "…/plugin/builtin"`）
3. **重新编译**
4. 后台 → **扩展** 启用（`bbs_kv.xiugo_plugins`）

### 已挂载的 Hook（核心会 Fire）

| Hook | 触发位置 | 事件类型 |
|------|----------|----------|
| `app.boot` | 启动 `plugin.Init` | nil |
| `page.render` | `renderPage` 写模板前 | `*PageRenderEvent`（Template, ExtraCSS/JS, FooterHTML） |
| `thread.viewed` | `Service.Thread` 浏览数+1 后 | `*ThreadViewedEvent` |
| `thread.created` | `Service.CreateThread` 成功 | `*ThreadCreatedEvent` |
| `reply.created` | `Service.Reply` 成功 | `*ReplyCreatedEvent` |
| `post.updated` | `Service.UpdatePost` 成功 | `*PostUpdatedEvent` |
| `post.deleted` | `Service.DeletePost` 成功 | `*PostDeletedEvent` |

`page.render` 仅作用于走 `renderPage` 的**前台**页。

### 没有 Fire 的地方

- 后台 `WriteTpl`、版主 modal 片段
- 需要新挂点时：在业务函数成功路径加一行  
  `plugin.Global().Fire(ctx, plugin.HookXxx, &plugin.XxxEvent{...})`

## 4. 推荐开发流程

### 只改颜色 / 间距

1. 复制 `resource/themes/hifini` → `resource/themes/my`
2. 改 CSS → 后台切换主题 → 强刷

### 改首页 HTML 结构

1. 复制 `resource/template/pages/home.html`  
   → `resource/themes/my/templates/pages/home.html`
2. 选中主题 `my` 后生效

### 加业务逻辑（不 fork 大段核心）

1. `builtin/myplugin.go` 注册 hook  
2. 若现有 hook 不够 → 在 `service.go` 对应成功返回前加 `Fire`  
3. 编译、后台启用  

### 打开开发辅助

后台启用扩展 **Developer Tools**：前台页 HTML 底部注释里会显示当前 `template=` 路径。

## 5. 明确不做

- Xiuno PHP 插件 / `{hook xxx.htm}` 兼容  
- 上传 zip 热插拔 PHP  
- 默认不使用 Go `.so` plugin  

## 6. 相关代码入口

| 能力 | 代码 |
|------|------|
| 主题扫描/解析 | `internal/theme/theme.go` |
| 插件管理器 | `internal/plugin/plugin.go` |
| 页面注入 | `internal/controller/web/controller.go` → `renderPage` |
| 业务 Fire | `internal/service/bbs/service.go` |
| 后台 UI | `/admin/themes`、`/admin/plugins` |
