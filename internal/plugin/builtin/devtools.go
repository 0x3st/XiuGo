package builtin

import (
	"context"
	"fmt"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&DevTools{})
	plugin.On("devtools", plugin.HookPageRender, onPageRender)
	plugin.On("devtools", plugin.HookThreadViewed, onThreadViewed)
	plugin.On("devtools", plugin.HookReplyCreated, onReplyCreated)
	plugin.On("devtools", plugin.HookThreadCreated, onThreadCreated)
}

// DevTools helps theme/plugin authors see which hooks and templates run.
// Enable it in Admin → 扩展. It injects an HTML comment before </body>.
type DevTools struct{}

func (d *DevTools) Meta() plugin.Info {
	return plugin.Info{
		ID:          "devtools",
		Name:        "Developer Tools",
		Version:     "1.0.0",
		Author:      "XiuGo",
		Description: "开发辅助：页脚注释显示当前模板名；浏览器控制台不输出。用于对照 URL↔模板与验证 Hook。",
		Builtin:     true,
	}
}

func (d *DevTools) OnEnable(ctx context.Context) error  { return nil }
func (d *DevTools) OnDisable(ctx context.Context) error { return nil }

func onPageRender(ctx context.Context, event any) error {
	e, ok := event.(*plugin.PageRenderEvent)
	if !ok || e == nil {
		return nil
	}
	e.FooterHTML += fmt.Sprintf(
		"\n<!-- xiugo-dev: template=%s theme-hook=page.render | edit: resource/template/%s or themes/<id>/templates/%s | hooks: see docs/dev-map.md -->\n",
		e.Template, e.Template, e.Template,
	)
	return nil
}

func onThreadViewed(ctx context.Context, event any) error {
	// Intentionally quiet — enable and use Fire count / future metrics.
	_ = event
	return nil
}

func onReplyCreated(ctx context.Context, event any) error {
	_ = event
	return nil
}

func onThreadCreated(ctx context.Context, event any) error {
	_ = event
	return nil
}
