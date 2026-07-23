package main

import (
	_ "github.com/0x3st/XiuGo/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"github.com/0x3st/XiuGo/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
