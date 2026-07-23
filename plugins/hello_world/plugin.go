// Package hello_world is a drop-in XiuGo plugin example.
//
// Install path: plugins/hello_world/
// Then: go generate ./internal/plugin/registry && go build
// Enable: Admin → 扩展 → hello_world
package hello_world

import (
	"context"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&Plugin{})
	plugin.On("hello_world", plugin.HookPageRender, func(ctx context.Context, event any) error {
		e, ok := event.(*plugin.PageRenderEvent)
		if !ok || e == nil {
			return nil
		}
		e.FooterHTML += `<!-- XiuGo drop-in plugin: hello_world -->`
		return nil
	})
}

// Plugin is a sample third-party style extension living under plugins/.
type Plugin struct{}

func (p *Plugin) Meta() plugin.Info {
	return plugin.Info{
		ID:          "hello_world",
		Name:        "Hello World (drop-in)",
		Version:     "1.0.0",
		Author:      "XiuGo Example",
		Description: "示例：把本目录拷进 plugins/ 后 go generate && 编译，后台启用即可在前台 HTML 注入注释。",
	}
}

func (p *Plugin) OnEnable(ctx context.Context) error  { return nil }
func (p *Plugin) OnDisable(ctx context.Context) error { return nil }
