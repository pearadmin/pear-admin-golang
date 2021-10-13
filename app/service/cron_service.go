package service

import (
	"github.com/robfig/cron/v3"
	"pear-admin-go/app/core/db"
	"pear-admin-go/app/core/log"
	"pear-admin-go/app/model"
	"time"
)

func InitCron() {
	log.Instance().Info("定时任务开启")
	c := cron.New()
	c.Schedule(cron.Every(5*time.Minute), cron.FuncJob(func() {
		DBReload()
	}))
}

func DBReload() {
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
