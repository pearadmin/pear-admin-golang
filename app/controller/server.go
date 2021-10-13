package controller

import (
	"github.com/cilidm/toolbox/gconv"
	"github.com/gin-gonic/gin"
	"net/http"
	"pear-admin-go/app/global/request"
	"pear-admin-go/app/global/response"
	"pear-admin-go/app/service"
	"pear-admin-go/app/util/validate"
)

func ServerList(c *gin.Context) {
	c.HTML(http.StatusOK, "server_list.html", nil)
}

func ServerJson(c *gin.Context) {
	var f request.TaskServerPage
	if err := c.ShouldBind(&f); err != nil {
		response.ErrorResp(c).SetMsg(validate.GetValidateError(err)).WriteJsonExit()
		return
	}
	data, count, err := service.ServerJson(f)
	if err != nil {
		response.SuccessResp(c).SetCode(0).SetMsg(err.Error()).SetCount(count).WriteJsonExit()
		return
	}
	response.SuccessResp(c).SetCode(0).SetCount(count).SetData(data).WriteJsonExit()
}

func ServerAdd(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "server_add.html", nil)
	} else {
		response.ErrorResp(c).SetMsg("请不要在演示服务器保存服务器信息").WriteJsonExit()
		return
		var f request.TaskServerForm
		if err := c.ShouldBind(&f); err != nil {
			response.ErrorResp(c).SetMsg(validate.GetValidateError(err)).WriteJsonExit()
			return
		}
		err := service.ServerAdd(f)
		if err != nil {
			response.ErrorResp(c).SetMsg(err.Error()).WriteJsonExit()
			return
		}
		response.SuccessResp(c).WriteJsonExit()
		return
	}
}

func ServerEdit(c *gin.Context) {
	if c.Request.Method == "GET" {
		id := c.Query("id")
		s, _ := service.FindServerById(gconv.Int(id))
		c.HTML(http.StatusOK, "server_edit.html", gin.H{"s": s})
	} else {
		var f request.TaskServerForm
		if err := c.ShouldBind(&f); err != nil {
			response.ErrorResp(c).SetMsg(validate.GetValidateError(err)).WriteJsonExit()
			return
		}
		err := service.ServerEdit(f)
		if err != nil {
			response.ErrorResp(c).SetMsg(err.Error()).WriteJsonExit()
			return
		}
		response.SuccessResp(c).WriteJsonExit()
		return
	}
}

func ServerDel(c *gin.Context) {
	id := c.PostForm("id")
	err := service.ServerDel(gconv.Int(id))
	if err != nil {
		response.ErrorResp(c).SetMsg(err.Error()).WriteJsonExit()
		return
	}
	response.SuccessResp(c).WriteJsonExit()
	return
}
