# 单 XiuGo 进程 部署：可清理清单

目标：**生产只跑 Go**；PHP 源码仅对照，不并行对外服务。

## 已完成（本轮）

| 项 | 处理 |
|----|------|
| 写路径 `invalidatePHPCache(forumlist/grouplist/thread_top_list)` | **已删除**（Go 不读这些键） |
| `invalidatePHPState` | **已删除** |
| 维护页语义 | 改为「重建运行数据 + 清 upload/tmp」，不再强调 PHP 编译缓存 |
| README | 去掉「并行运行」表述 |

## 保留（与是否并行无关）

| 项 | 原因 |
|----|------|
| `bbs_*` 表读写 | 业务数据 |
| `bbs_cache` + runtime | Go 仍存/读 `bbs_runtime` |
| `phpRoot` / 读 `conf.php` | 站名、SMTP、时区、附件目录规则等仍从原配置目录读取 |
| `upload/` 共用路径 | 附件与头像文件 |
| 密码 MD5+salt | 老用户登录 |
| 维护工具入口 | 运维重建 runtime / 清临时文件 |

## 后续可继续清理（未在本轮删除）

1. **配置迁出 `conf.php`**：站名/SMTP/runlevel 进 `manifest/config/config.yaml`，减少对 `phpRoot` 依赖  
2. **插件管理页**：若永不做插件体系，可隐藏 `/admin/?plugin*`  
3. **admin 双轨**：现代 `/admin/*` 与 `?xxx.htm` 兼容层二选一收敛  
4. **台账验收标准**：由「PHP 交叉读取」改为「Go 单端 + 库表断言」  
5. **cron 路径文案**：`cleanupOriginalTempAttachments` 命名可改为中性  
6. **`SyncPHPRuntime` 命名**：可改名为 `SyncRuntime`（行为可先不动）

## 不要误删

- `phpCacheKeys` / `loadPHPRuntime` / `savePHPRuntime` / `SyncPHPRuntime`：Go 仍用  
- 用户/帖子/附件删除与计数逻辑  
- 前台 UI 对齐原版的模板与静态资源  

## 开发建议

- 日常只启动 `xiuno-go/start-local.sh`（8081）  
- 需要肉眼对照 DOM 时再临时开 PHP `8080`  
