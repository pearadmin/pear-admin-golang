package service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"pear-admin-go/app/core/log"
	dao "pear-admin-go/app/dao"
	"pear-admin-go/app/util/clientIP"

	e2 "pear-admin-go/app/global/e"
	"pear-admin-go/app/global/request"
	"pear-admin-go/app/model"
	"time"
)

func CreateOperLog(c *gin.Context, f model.OperForm) error {
	var operName string
	user := GetProfile(c)
	if user == nil {
		operName = ""
	} else {
		operName = user.LoginName
	}
	outJson, _ := json.Marshal(f.OutContent)
	var oper model.OperLog
	oper.Title = f.Title
	oper.OperParam = f.InContent
	oper.JsonResult = string(outJson)
	oper.BusinessType = int(f.OutContent.Type)
	oper.OperatorType = 1 // 操作类别（0其它 1后台用户 2手机端用户）
	if f.OutContent.Code == 0 {
		oper.Status = 0
	} else {
		oper.Status = 1
	}
	oper.OperName = operName
	oper.RequestMethod = c.Request.Method
	oper.OperUrl = c.Request.URL.Path
	oper.Method = c.Request.Method
	oper.OperIp = clientIP.GetIp(c.Request)
	//oper.OperLocation = ip.GetCityByIp(oper.OperIp)
	oper.OperLocation, _ = clientIP.GetCityByIP(oper.OperIp)
	oper.ErrorMsg = f.ErrorMsg
	oper.OperTime = time.Now().Format(e2.TimeFormat)
	if err := dao.NewOperLogDaoImpl().Insert(oper); err != nil {
		return err
	}
	return nil
}

func OperLogListJsonService(f request.LayerListForm) (count int, list []model.OperLog, err error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 10
	}
	list, count, err = dao.NewOperLogDaoImpl().FindByPage(f.Page, f.Limit)
	if err != nil {
		log.Instance().Error("LoginInfoListJsonService.FindByPage:" + err.Error())
		return 0, nil, err
	}
	for k, _ := range list {
		list[k].OperIp = "*.*.*.*"
	}
	return count, list, nil
}
