package builtin

import (
	"context"
	"strings"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&CreditsAdmin{})
	plugin.On("credits_admin", plugin.HookAdminRender, onAdminRenderCredits)
	plugin.On("credits_admin", plugin.HookPageRender, onPageRenderCredits)
}

// CreditsAdmin is the credits feature pack:
// - front profile /my shows 积分
// - admin user list shows credits column
// - (awards wired in core when this plugin is enabled)
type CreditsAdmin struct{}

func (c *CreditsAdmin) Meta() plugin.Info {
	return plugin.Info{
		ID:          "credits_admin",
		Name:        "积分系统",
		Version:     "2.0.0",
		Author:      "XiuGo",
		Description: "积分插件：个人中心/用户资料显示积分；后台用户列表显示积分；发主题+2、回帖+1。管理：/admin/credits",
		Builtin:     true,
	}
}

func (c *CreditsAdmin) OnEnable(ctx context.Context) error  { return nil }
func (c *CreditsAdmin) OnDisable(ctx context.Context) error { return nil }

func onAdminRenderCredits(ctx context.Context, event any) error {
	e, ok := event.(*plugin.AdminRenderEvent)
	if !ok || e == nil {
		return nil
	}
	if e.Template != "admin/users.html" && !strings.Contains(e.Template, "user_list.html") && e.Template != "admin/credits.html" {
		return nil
	}
	if e.Params == nil {
		e.Params = map[string]any{}
	}
	e.Params["ShowUserCredits"] = true
	return nil
}

func onPageRenderCredits(ctx context.Context, event any) error {
	e, ok := event.(*plugin.PageRenderEvent)
	if !ok || e == nil {
		return nil
	}
	if e.Template != "pages/user_profile.html" {
		return nil
	}
	// renderPage merges Plugin params only Footer/CSS — ShowUserCredits must be in main params.
	// PageRenderEvent.Params is separate; controller should merge flags from event.Params too.
	if e.Params == nil {
		e.Params = map[string]any{}
	}
	e.Params["ShowUserCredits"] = true
	return nil
}
