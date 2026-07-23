package builtin

import (
	"context"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&Hello{})
	plugin.On("hello", plugin.HookPageRender, func(ctx context.Context, event any) error {
		e, ok := event.(*plugin.PageRenderEvent)
		if !ok || e == nil {
			return nil
		}
		// Tiny visible marker when enabled — proves hook pipeline works.
		e.FooterHTML += `<!-- XiuGo plugin:hello enabled -->`
		return nil
	})
}

// Hello is a sample Go-native plugin. Not related to Xiuno PHP plugins.
type Hello struct{}

func (h *Hello) Meta() plugin.Info {
	return plugin.Info{
		ID:          "hello",
		Name:        "Hello Extension",
		Version:     "1.0.0",
		Author:      "XiuGo",
		Description: "示例扩展：启用后在页面 HTML 注入注释标记，验证 Hook 管线。",
		Builtin:     true,
	}
}

func (h *Hello) OnEnable(ctx context.Context) error  { return nil }
func (h *Hello) OnDisable(ctx context.Context) error { return nil }
