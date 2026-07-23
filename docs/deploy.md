# XiuGo 部署说明（5.0.x）

## 要求

- Go 1.23+
- MySQL 5.7+ / 8.x（Xiuno 兼容库表 `bbs_*`）
- 可写上传目录

## 配置

```sh
cp manifest/config/config.example.yaml manifest/config/config.yaml
```

必填：

- `database.default.link`
- `xiuno.uploadPath`（头像、附件、版块图标）

可选：

- `xiuno.phpRoot`：仅用于从旧 `conf.php` / `smtp.conf.php` **首次导入**；可留空
- `xiuno.sitename` / `runlevel` / `timezone` 等：当库中尚无 `bbs_kv` 配置时的默认值

站点与 SMTP 运行时保存在：

- `bbs_kv.k = xiugo_site`
- `bbs_kv.k = xiugo_smtp`

## 启动

```sh
./start-local.sh
# 或
go run .
```

- 站点：`http://HOST:8081/`
- 后台：`http://HOST:8081/admin/`（管理员需二次验证密码）
- 设置：`/admin/settings`、`/admin/settings/smtp`
- 维护：`/admin/maintenance`

## 后台路径

| 功能 | 路径 |
|------|------|
| 首页 | `/admin/` |
| 二次登录 | `/admin/login` |
| 站点设置 | `/admin/settings` |
| SMTP | `/admin/settings/smtp` |
| 主题 | `/admin/threads` |
| 用户 | `/admin/users` |
| 版块 | `/admin/forums` |
| 用户组 | `/admin/groups` |
| 维护 | `/admin/maintenance` |

旧式 `/admin/?setting-base.htm` 等查询串入口仍可能可用，新部署请只用上表路径。

## 安全

- 生产必须 HTTPS
- 勿将 `config.yaml` 提交到仓库
- 浏览器 MD5 仅兼容旧协议，不能替代 TLS

## 分支

- `develop`：日常开发与测试
- `main`：相对稳定线
