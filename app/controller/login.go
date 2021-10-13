package controller

import (
	"github.com/cilidm/toolbox/gconv"
	"github.com/cilidm/toolbox/ip"
	pkg "github.com/cilidm/toolbox/str"
	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"net/http"
	"pear-admin-go/app/dao"
	"pear-admin-go/app/global/e"
	"pear-admin-go/app/global/request"
	"pear-admin-go/app/global/response"
	"pear-admin-go/app/model"
	"pear-admin-go/app/service"
	"pear-admin-go/app/util/captcha"
	"pear-admin-go/app/util/clientIP"
	"pear-admin-go/app/util/validate"
	"strings"
	"time"
)

func Login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func LoginHandler(c *gin.Context) {
	var req request.LoginForm
	if err := c.ShouldBind(&req); err != nil {
		response.ErrorResp(c).SetMsg(validate.GetValidateError(err)).WriteJsonExit()
		return
	}
	isLock := service.CheckLock(req.UserName)
	if isLock {
		response.ErrorResp(c).SetMsg("密码错误次数超限，账号已锁定,请稍后再试").SetType(model.OperOther).Log(e.LoginHandler, req).WriteJsonExit()
		return
	}
	userAgent := c.Request.Header.Get("User-Agent")
	ua := user_agent.New(userAgent)
	ub, _ := ua.Browser()

	var info model.LoginInfo
	info.LoginName = req.UserName
	info.IpAddr = clientIP.GetIp(c.Request)
	info.Os = ua.OS()
	info.Browser = ub
	info.LoginTime = time.Now()
	info.LoginLocation = ip.GetCityByIp(clientIP.GetIp(c.Request))

	if sid, err := service.SignIn(req.UserName, req.Password, c); err != nil {
		errNums := service.SetPwdErrNum(req.UserName)
		having := e.MaxErrNum - errNums
		info.Msg = "账号或密码错误"
		info.Status = "0"
		err = dao.NewLoginInfoImpl().Insert(info)
		if err != nil {
			response.ErrorResp(c).SetMsg(err.Error()).SetType(model.OperOther).Log(e.LoginHandler, req).WriteJsonExit()
			return
		}
		if having > 0 {
			response.ErrorResp(c).SetMsg("账号或密码不正确,还有"+gconv.String(having)+"次之后账号将锁定").SetType(model.OperOther).Log(e.LoginHandler, req).WriteJsonExit()
			return
		} else {
			response.ErrorResp(c).SetMsg("密码错误次数超限，账号已锁定,请稍后再试").SetType(model.OperOther).Log(e.LoginHandler, req).WriteJsonExit()
			return
		}
	} else {
		var online model.AdminOnline
		err = pkg.CopyFields(&online, info)
		if err != nil {
			response.ErrorResp(c).SetMsg(err.Error()).SetType(model.OperOther).Log(e.LoginHandler, req).WriteJsonExit()
			return
		}
		online.SessionID = sid
		online.Status = "on_line"
		online.ExpireTime = 1440
		online.StartTimestamp = time.Now().Unix()
		online.LastAccessTime = time.Now().Format(e.TimeFormat)
		dao.NewAdminOnlineDaoImpl().Delete(sid)
		dao.NewAdminOnlineDaoImpl().Insert(online)
		service.RemovePwdErrNum(req.UserName)

		info.Msg = "登陆成功"
		info.Status = "1"
		err := dao.NewLoginInfoImpl().Insert(info)
		if err != nil {
			response.ErrorResp(c).SetMsg(err.Error()).SetType(model.OperOther).Log(e.LoginHandler, nil).WriteJsonExit()
		}
		response.SuccessResp(c).SetMsg("登陆成功").SetType(model.OperOther).Log(e.LoginHandler, nil).WriteJsonExit()
	}
}

// 注销
func Logout(c *gin.Context) {
	if service.IsSignedIn(c) {
		service.SignOut(c)
	}
	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}

func NotFound(c *gin.Context) {
	c.HTML(http.StatusOK, "not_found.html", gin.H{})
}

func GetCaptcha(c *gin.Context) {
	id, b64s, err := captcha.CaptMake()
	if err != nil {
		response.ErrorResp(c).SetMsg(err.Error()).WriteJsonExit()
		return
	}
	response.SuccessResp(c).SetData(response.CaptchaResponse{CaptchaId: id, PicPath: b64s}).WriteJsonExit()
}

func CaptchaVerify(c *gin.Context) {
	id := c.PostForm("id")
	capt := c.PostForm("capt")
	if id == "" || capt == "" {
		response.ErrorResp(c).SetMsg("请填写完整信息").WriteJsonExit()
		return
	}

	if captcha.CaptVerify(id, strings.ToLower(capt)) == true {
		response.SuccessResp(c).WriteJsonExit()
	} else {
		response.ErrorResp(c).SetMsg("验证码有误，请重新输入").WriteJsonExit()
	}
	return
}
