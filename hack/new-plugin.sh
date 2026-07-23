#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ID="${1:-}"
if [[ -z "$ID" ]]; then
  echo "用法: $0 <plugin_id>"
  echo "例:   $0 my_stats"
  exit 1
fi
if [[ ! "$ID" =~ ^[a-z][a-z0-9_]*$ ]]; then
  echo "plugin_id 请用小写字母开头，仅 a-z0-9_"
  exit 1
fi
DIR="$ROOT/plugins/$ID"
if [[ -e "$DIR" ]]; then
  echo "已存在: $DIR"
  exit 1
fi
mkdir -p "$DIR"
cat > "$DIR/plugin.go" <<EOF
package ${ID}

import (
	"context"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&Plugin{})
	// 按需注册 hook，例如：
	// plugin.On("${ID}", plugin.HookPageRender, func(ctx context.Context, event any) error {
	// 	return nil
	// })
}

// Plugin 是 drop-in 扩展，放在 plugins/${ID}/ 下。
type Plugin struct{}

func (p *Plugin) Meta() plugin.Info {
	return plugin.Info{
		ID:          "${ID}",
		Name:        "${ID}",
		Version:     "0.1.0",
		Author:      "",
		Description: "TODO: 描述插件做什么",
	}
}

func (p *Plugin) OnEnable(ctx context.Context) error  { return nil }
func (p *Plugin) OnDisable(ctx context.Context) error { return nil }
EOF
echo "已创建 $DIR/plugin.go"
echo "下一步:"
echo "  go generate ./internal/plugin/registry"
echo "  go build -o xiugo ."
echo "  后台 → 扩展 → 启用 ${ID}"
