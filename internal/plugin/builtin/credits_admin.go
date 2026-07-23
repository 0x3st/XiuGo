package builtin

import (
	"context"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&CreditsAdmin{})
	plugin.On("credits_admin", plugin.HookAdminRender, onAdminRenderCredits)
}

// CreditsAdmin shows bbs_user.credits on the modern admin user list when enabled.
// This is a Go-native plugin demo — not a Xiuno PHP plugin.
type CreditsAdmin struct{}

func (c *CreditsAdmin) Meta() plugin.Info {
	return plugin.Info{
		ID:          "credits_admin",
		Name:        "User Credits (Admin)",
		Version:     "1.0.0",
		Author:      "XiuGo",
		Description: "后台用户列表显示积分 credits。启用后访问 /admin/users 可见「积分」列。",
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
	if e.Template != "admin/users.html" {
		return nil
	}
	if e.Params == nil {
		e.Params = map[string]any{}
	}
	// Template shows the credits column when this flag is set.
	e.Params["ShowUserCredits"] = true
	return nil
}
