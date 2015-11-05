package server

import (
	"fmt"
	"log"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/nzai/stockrecorder/config"
)

//	启动服务
func Start() {

	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	//	注册路由
	registerRoute(e)

	log.Print("启动Http服务")
	// Start server
	e.Run(fmt.Sprintf(":%d", config.Get().Port))
}
