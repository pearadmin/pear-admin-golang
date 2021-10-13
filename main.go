package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/cilidm/toolbox/gconv"
	"net/http"
	"os"
	"os/signal"
	"pear-admin-go/app/core/config"
	"pear-admin-go/app/core/db"
	"pear-admin-go/app/core/log"
	"pear-admin-go/app/core/redis"
	"pear-admin-go/app/router"
	"pear-admin-go/app/service"
	"pear-admin-go/app/util/validate"
	"syscall"
	"time"
)

//go:embed template
var templateFs embed.FS

//go:embed static
var staticFs embed.FS

func main() {
	if err := validate.InitTrans("zh"); err != nil {
		fmt.Println("init trans failed, err:", err)
	}

	config.InitConfig("./config.toml")
	log.InitLog()
	db.InitConn()

	redis.InitRedis()
	r := router.InitRouter(staticFs, templateFs)
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Instance().App.HttpPort),
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf(`	欢迎使用 Pear Admin Go
	程序运行地址:http://127.0.0.1:%s
`, gconv.String(config.Instance().App.HttpPort))
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Instance().Error(err.Error())
			os.Exit(0)
		}
	}()
	go service.InitCron()
	shutDown(s)
}

func shutDown(s *http.Server) {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutdown Server ...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := db.Instance().Close()
	if err != nil {
		log.Instance().Fatal("Close DB error:" + err.Error())
	}

	if err := s.Shutdown(ctx); err != nil {
		log.Instance().Fatal("Server Shutdown:" + err.Error())
	}
	fmt.Println("Server exiting")
}
