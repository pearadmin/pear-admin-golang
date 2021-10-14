package service

import (
	"github.com/robfig/cron/v3"
	"pear-admin-go/app/core/db"
	"pear-admin-go/app/core/log"
	"pear-admin-go/app/model"
	"time"
)

func InitCron() {
	c := cron.New()
	c.Schedule(cron.Every(5*time.Minute), cron.FuncJob(func() {
		DBReload()
	}))
	c.Start()
}

func DBReload() {
	log.Instance().Info("开始执行定时任务")
	db.Instance().DropTableIfExists(
		&model.Admin{},
		&model.AdminOnline{},
		&model.Auth{},
		&model.Role{},
		&model.RoleAuth{},
		&model.SysConf{},
		&model.PearConfig{},
		&model.Task{},
		&model.TaskServer{},
	)
	db.InitConn()
}
