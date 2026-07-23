package cmd

import (
	"context"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/0x3st/XiuGo/internal/controller/web"
	"github.com/0x3st/XiuGo/internal/service/bbs"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			if err = bbs.ApplyOriginalTimezone(ctx); err != nil {
				return err
			}
			if err = bbs.StartOriginalMaintenance(ctx); err != nil {
				return err
			}
			var (
				s          = g.Server()
				controller = web.New()
			)
			if uploadPath := g.Cfg().MustGet(ctx, "xiuno.uploadPath").String(); uploadPath != "" {
				s.AddStaticPath("/upload", uploadPath)
			}
			s.Group("/", func(group *ghttp.RouterGroup) {
				controller.Bind(group)
			})
			s.Run()
			return nil
		},
	}
)
