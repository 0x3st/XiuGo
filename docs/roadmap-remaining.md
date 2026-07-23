# XiuGo 还可补充的能力（P0/P1 之后）

## 已在本阶段补强

- 列表/主题/用户主题分页
- 版块 orderby + page
- 编辑首帖移动板块
- 主题 `?keyword=` 标题高亮
- 列表按用户组可读板块过滤
- 首页 order_default
- SMTP 测试发送入口
- 台账同步（头像、move、分页等）

## 仍值得做（产品）

| 优先级 | 项 | 说明 |
|--------|-----|------|
| 高 | ~~配置迁出 conf.php~~ | **站点/SMTP 已入 bbs_kv**；uploadPath 在 yaml；phpRoot 仅可选导入 |
| 高 | 后台路由单一化 | 只保留 `/admin/*` 或只保留 `?*.htm` |
| 中 | 全站搜索页 | 原版核心几乎无独立搜索；可做 subject LIKE |
| 中 | 主题楼层锚点与“高级回帖”附件 | 体验打磨 |
| 中 | 多语言 UI | setting 有 lang 选项，前台文案仍中文硬编码 |
| 低 | ~~PHP 插件页~~ | **已移除**（XiuGo 不加载 Xiuno 插件） |
| 低 | SyncRuntime 命名清理 | 纯重构 |
| 不做 | ~~插件市场~~（已砍入口）/ synlogin / phpinfo |
| 可选 | 自研 Go 扩展点（见下文） | 与 Xiuno 插件 **不兼容** |

## 运维

- 维护工具可选隐藏（非必需）
- 生产 HTTPS + 真实 SMTP


## 关于 Go / GoFrame 插件生态

- **不能** 直接跑 Xiuno 的 PHP 插件（Hook、Overwrite、`install.php`）。
- **可以** 做 XiuGo 自己的扩展方式，例如：
  1. **进程内接口注册**：定义 `OnThreadCreate` 等 hook，业务用 `Register(...)` 挂实现（最常见、最好维护）。
  2. **GoFrame 模块化**：按 `internal/module/xxx` 分包，编译进主程序（不是热插拔）。
  3. **gRPC / HashiCorp go-plugin**：独立进程插件，适合隔离与多语言，复杂度高。
  4. **官方 `plugin` 包（.so）**：平台相关、版本敏感，生产论坛一般不推荐。

GoFrame **没有** 现成的「Xiuno 式插件市场」；要有生态，需要 XiuGo 自己定义 API 与分发方式。
