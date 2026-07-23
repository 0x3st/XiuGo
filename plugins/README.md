# 第三方 / 外置插件目录（drop-in）

把插件放在这里，**重新生成注册表并编译**后，即可在后台「扩展」里启用。

> 仍然是编译进主程序，不是运行时上传 zip。  
> 目标：开发完一个插件后，**拷贝目录 + 两条命令** 就能用。

## 快速接入（3 步）

```bash
# 1. 放入插件（目录名建议 = 插件 ID）
cp -R /path/to/my_plugin plugins/my_plugin

# 2. 生成 import 注册表
go generate ./internal/plugin/registry
# 或: make plugins

# 3. 编译并运行
go build -o xiugo .
./xiugo
```

后台 → **扩展** → 启用你的插件。

## 插件目录结构

```text
plugins/
  my_plugin/
    plugin.go          # package my_plugin（包名任意，目录名用于 import 路径）
    ... 其他 .go 文件
```

`plugin.go` 最小示例：

```go
package my_plugin

import (
	"context"
	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&P{})
	// plugin.On("my_plugin", plugin.HookPageRender, ...)
}

type P struct{}

func (p *P) Meta() plugin.Info {
	return plugin.Info{
		ID: "my_plugin", Name: "我的插件", Version: "1.0.0",
		Description: "...",
	}
}
func (p *P) OnEnable(context.Context) error  { return nil }
func (p *P) OnDisable(context.Context) error { return nil }
```

## 约定

| 项 | 说明 |
|----|------|
| 目录名 | 用于 `import github.com/0x3st/XiuGo/plugins/<dir>` |
| 插件 ID | `Meta().ID`，后台启用用；建议与目录名一致 |
| 内置插件 | 仍在 `internal/plugin/builtin`（hello、devtools、credits_admin） |
| PHP 插件 | **不支持** |

## 脚手架

```bash
./hack/new-plugin.sh my_plugin
go generate ./internal/plugin/registry
go build .
```

## 别人给你一个插件时

对方交付整个 `plugins/xxx` 文件夹即可（纯 Go 源码）。  
你拷贝进本仓库的 `plugins/`，执行 generate + build，无需改 `cmd` 手写 import。
