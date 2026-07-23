package builtin

import (
	"context"
	"strings"

	"github.com/0x3st/XiuGo/internal/plugin"
)

func init() {
	plugin.Register(&CreditsAdmin{})
	plugin.On("credits_admin", plugin.HookAdminRender, onAdminRenderCredits)
}

// CreditsAdmin shows bbs_user.credits on admin user lists when enabled.
type CreditsAdmin struct{}

func (c *CreditsAdmin) Meta() plugin.Info {
	return plugin.Info{
		ID:          "credits_admin",
		Name:        "User Credits (Admin)",
		Version:     "1.1.0",
		Author:      "XiuGo",
		Description: "后台用户列表显示积分。启用后打开 /admin/users（或兼容用户列表）可见「积分」列。数值为 bbs_user.credits，未做加分规则时多为 0。",
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
	// Modern + compat user list templates
	if e.Template != "admin/users.html" && !strings.Contains(e.Template, "user_list.html") {
		return nil
	}
	if e.Params == nil {
		e.Params = map[string]any{}
	}
	e.Params["ShowUserCredits"] = true
	return nil
}
