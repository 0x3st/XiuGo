package cmd

import (
	"context"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/0x3st/XiuGo/internal/controller/web"
	_ "github.com/0x3st/XiuGo/internal/plugin/builtin"
	"github.com/0x3st/XiuGo/internal/plugin"
	"github.com/0x3st/XiuGo/internal/service/bbs"
	"github.com/0x3st/XiuGo/internal/theme"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start XiuGo http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			if err = bbs.ApplyOriginalTimezone(ctx); err != nil {
				return err
			}
			if err = bbs.StartOriginalMaintenance(ctx); err != nil {
				return err
			}
			svc := bbs.New()
			if err = theme.Init(ctx, "resource/themes", svc.LoadActiveThemeID); err != nil {
				return err
			}
			if err = plugin.Init(ctx, svc.LoadPluginEnabledMap(ctx)); err != nil {
				return err
			}
			var (
				s          = g.Server()
				controller = web.New()
			)
			if uploadPath := g.Cfg().MustGet(ctx, "xiuno.uploadPath").String(); uploadPath != "" {
				s.AddStaticPath("/upload", uploadPath)
			}
			// Theme static files: /themes/<id>/...
			if root := theme.Global().Root(); root != "" {
				if st, err := os.Stat(root); err == nil && st.IsDir() {
					s.AddStaticPath("/themes", root)
				}
			}
			s.Group("/", func(group *ghttp.RouterGroup) {
				controller.Bind(group)
			})
			s.Run()
			return nil
		},
	}
)
