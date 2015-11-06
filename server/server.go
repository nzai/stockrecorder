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
	//	e.Use(mw.Logger())
	//	e.Use(mw.Recover())
	e.Use(mw.Gzip())

	//	注册路由
	registerRoute(e)

	log.Printf("启动Http服务,端口:%d", config.Get().Port)
	// Start server
	e.Run(fmt.Sprintf(":%d", config.Get().Port))
}
