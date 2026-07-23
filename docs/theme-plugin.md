# XiuGo Theme & Plugin

XiuGo 自有扩展体系，**不兼容** Xiuno PHP 主题/插件（无 Hook/Overwrite/`install.php`）。

## Theme

- 目录：`resource/themes/<id>/`
- 清单：`theme.json`（name/version/css 等）
- 样式：通常 `css/theme.css`，通过 `/themes/<id>/css/theme.css` 访问
- 激活：后台 **主题**，持久化键 `bbs_kv.xiugo_theme`

内置：

| ID | 说明 |
|----|------|
| `default` | 默认 Xiuno 风格（仅 stock CSS） |
| `hifini` | 灰蓝背景 + 蓝顶栏列表，外观灵感来自常见音乐向 Xiuno 站（非官方复制） |

## Plugin（Go 原生）

- 实现 `plugin.Plugin`，在 `init()` 里 `plugin.Register` + `plugin.On(hook, handler)`
- 示例：`internal/plugin/builtin/hello.go`
- 启用状态：`bbs_kv.xiugo_plugins`
- 后台 **扩展** 开关

常用 Hook：

| Hook | 事件 |
|------|------|
| `app.boot` | 启动 |
| `page.render` | 渲染前台页（ExtraCSS/FooterHTML） |
| `thread.viewed` | 看帖浏览数更新后 |
| `thread.created` | 发主题成功 |
| `reply.created` | 回帖成功 |
| `post.updated` / `post.deleted` | 编辑/删除帖 |

热插拔 `.so` / PHP 插件目录 **不在范围**。扩展以编译进主程序的 Go 包为主。

主题可覆盖 HTML：`resource/themes/<id>/templates/` 下镜像 `resource/template/` 路径。


更完整的 URL↔模板↔Hook 表见 [dev-map.md](dev-map.md)。
