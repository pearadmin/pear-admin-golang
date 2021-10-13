package service

import (
	"encoding/json"
	"errors"
	"github.com/cilidm/toolbox/gconv"
	pkg "github.com/cilidm/toolbox/str"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pear-admin-go/app/core/cache"
	"pear-admin-go/app/core/config"
	"pear-admin-go/app/core/log"
	"pear-admin-go/app/dao"
	e2 "pear-admin-go/app/global/e"
	"pear-admin-go/app/global/request"
	"pear-admin-go/app/model"
	"pear-admin-go/app/util/session"
	"strings"
	"sync"
	"time"
)

func BuildFilter(r request.AdminForm) []interface{} {
	filters := make([]interface{}, 0)
	if r.ID != "" {
		filters = append(filters, "id = ?", r.ID)
	}
	if r.LoginName != "" {
		filters = append(filters, "login_name LIKE ?", "%"+r.LoginName+"%")
	}
	if r.RealName != "" {
		filters = append(filters, "real_name LIKE ?", "%"+r.RealName+"%")
	}
	if r.Phone != "" {
		filters = append(filters, "phone LIKE ?", "%"+r.Phone+"%")
	}
	if r.Email != "" {
		filters = append(filters, "email LIKE ?", "%"+r.Email+"%")
	}
	return filters
}

func AdminListJsonService(f request.AdminForm) (count int, data []map[string]interface{}, err error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 10
	}
	filters := BuildFilter(f)
	list, count, err := dao.NewAdminDaoImpl().FindByPage(f.Page, f.Limit, filters...)
	if err != nil {
		log.Instance().Error("AdminListJsonService.FindByPage", zap.Error(err))
		return count, data, err
	}
	for _, v := range list {
		data = append(data, pkg.Struct2MapByTag(model.AdminShow{
			ID:        v.ID,
			LoginName: v.LoginName,
			RealName:  v.RealName,
			RoleIds:   v.RoleIds,
			Level:     v.Level,
			Phone:     v.Phone,
			Email:     v.Email,
			Avatar:    v.Avatar,
			Status:    v.Status,
			LastIp:    v.LastIp,
			LastLogin: v.LastLogin,
		}, "json"))
	}
	return count, data, nil
}

func CreateAdminService(r request.AdminAddForm, uid int) error {
	if CheckLoginName(r.LoginName) {
		return errors.New("用户名已存在")
	}
	pwd, salt := pkg.SetPassword(e2.DefaultSaltLen, r.Password)
	_, err := dao.NewAdminDaoImpl().Insert(model.Admin{
		LoginName: r.LoginName,
		Password:  pwd,
		Salt:      salt,
		RealName:  r.RealName,
		RoleIds:   r.RoleIds,
		Level:     1,
		Phone:     r.Phone,
		Email:     r.Email,
		Status:    r.Status,
		CreateId:  uid,
		UpdateId:  uid,
		LastLogin: time.Now().Format(e2.TimeFormat),
	})
	if err != nil {
		log.Instance().Error("CreateAdminService.Insert", zap.Error(err))
		return err
	}
	return nil
}

func UpdateAdminStatus(id, status string) error {
	attr := make(map[string]interface{})
	attr["status"] = status
	admin, err := dao.NewAdminDaoImpl().FindAdmin("id = ?", id)
	if err != nil {
		log.Instance().Error("UpdateAdminStatus.FindAdmin", zap.Error(err))
		return err
	}
	if admin.ID < 1 {
		return errors.New("未找到用户")
	}
	err = dao.NewAdminDaoImpl().Update(admin, attr)
	return err
}

func UpdateAdminAttrService(f request.AdminEditForm) error {
	attr := make(map[string]interface{})
	attr["status"] = f.Status
	attr["role_ids"] = f.RoleIds
	if f.LoginName != "" {
		attr["login_name"] = f.LoginName
	}
	if f.RealName != "" {
		attr["real_name"] = f.RealName
	}
	if f.Phone != "" {
		attr["phone"] = f.Phone
	}
	if f.Email != "" {
		attr["email"] = f.Email
	}
	admin, err := dao.NewAdminDaoImpl().FindAdmin("id = ?", f.ID)
	if err != nil {
		log.Instance().Error("UpdateAdminAttrService.FindAdmin", zap.Error(err))
		return err
	}
	if admin.ID < 1 {
		return errors.New("未找到用户")
	}
	err = dao.NewAdminDaoImpl().Update(admin, attr)
	if err != nil {
		log.Instance().Error("UpdateAdminAttrService.Update", zap.Error(err))
		return err
	}
	return nil
}

func AdminEditService(uid string) (show model.AdminShow, rolesShow []model.RoleEditShow, err error) {
	admin, err := dao.NewAdminDaoImpl().FindAdmin("id = ?", uid)
	if err != nil {
		log.Instance().Error("AdminEditService.FindAdmin", zap.Error(err))
		return show, rolesShow, err
	}
	err = pkg.CopyFields(&show, admin)
	if err != nil {
		log.Instance().Error("AdminEditService.CopyFields", zap.Error(err))
		return show, rolesShow, err
	}
	roles, err := dao.NewRoleDaoImpl().FindRoles("status = ?", "1") // 查找全部的分组
	if err != nil {
		log.Instance().Error("AdminEditService.FindRoles", zap.Error(err))
		return show, rolesShow, err
	}
	check := strings.Split(admin.RoleIds, ",")
	for _, v := range roles {
		checked := 0
		if pkg.IsContain(check, gconv.String(v.ID)) {
			checked = 1
		}
		rolesShow = append(rolesShow, model.RoleEditShow{
			ID:       gconv.Int(v.ID),
			RoleName: v.RoleName,
			Status:   v.Status,
			Checked:  checked,
		})
	}
	return show, rolesShow, nil
}

func AdminAddHandlerService(roleIds, status string, f request.AdminAddForm, c *gin.Context) error {
	if pkg.IsEmail([]byte(f.Email)) == false {
		return errors.New("邮箱格式不符合规范")
	}
	f.RoleIds = roleIds
	f.Status = gconv.Int(status)
	if err := CreateAdminService(f, GetUid(c)); err != nil {
		log.Instance().Error("AdminAddHandlerService.CreateAdminService", zap.Error(err))
		return err
	}
	return nil
}

func AdminDeleteService(uid string, c *gin.Context) error {
	admin, err := dao.NewAdminDaoImpl().FindAdmin("id = ?", uid)
	if err != nil {
		log.Instance().Error("AdminDeleteService.FindAdmin", zap.Error(err))
		return err
	}
	if admin.ID == 1 || admin.Level == 99 {
		return errors.New("超级管理员不能删除")
	}
	userId := GetUid(c)
	if admin.ID == gconv.Uint(userId) {
		return errors.New("不能删除本人账号")
	}
	if err := dao.NewAdminDaoImpl().Delete(gconv.Int(uid)); err != nil {
		log.Instance().Error("AdminDeleteService.Delete", zap.Error(err))
		return err
	}
	return nil
}

// 判断是否是超级管理员
func IsAdmin(user *model.Admin) bool {
	return user.ID == 1 && user.Level == 99 // 只有id = 1 && level = 99 的超级管理员
}

func GetUid(c *gin.Context) int {
	uid := session.Get(c, e2.Auth)
	if uid != nil {
		return gconv.Int(uid)
	}
	return 0
}

// 判断用户是否已经登录
func IsSignedIn(c *gin.Context) bool {
	uid := session.Get(c, e2.Auth)
	return uid != nil
}

// 用户登录，成功返回用户信息，否则返回nil; passport应当会md5值字符串
func SignIn(loginName, password string, c *gin.Context) (string, error) {
	//查询用户信息
	admin, err := dao.NewAdminDaoImpl().FindByName(loginName)
	if err != nil || (admin.Password != pkg.Md5([]byte(password+admin.Salt))) {
		return "", errors.New("账号或密码错误")
	}
	update := make(map[string]interface{})
	update["last_ip"] = c.ClientIP()
	update["last_login"] = time.Now()
	if err := dao.NewAdminDaoImpl().Update(admin, update); err != nil {
		return "", err
	}
	adminID, err := SaveUserToSession(admin, c)
	if err != nil {
		return "", err
	}
	return adminID, nil
}

var SessionList sync.Map

//保存用户信息到session
func SaveUserToSession(admin model.Admin, c *gin.Context) (string, error) {
	err := session.Del(c, e2.Auth)
	if err != nil {
		return "", err
	}
	err = session.Set(c, e2.Auth, admin.ID)
	if err != nil {
		log.Instance().Warn(err.Error())
		return "", err
	}
	tmp, _ := json.Marshal(admin)
	err = session.Set(c, e2.AdminInfo, string(tmp))
	if err != nil {
		log.Instance().Warn(err.Error())
		return "", err
	}
	SessionList.Store(admin.ID, c)
	return gconv.String(admin.ID), nil
}

//清空用户菜单缓存
func ClearMenuCache(user *model.Admin) {
	if IsAdmin(user) {
		cache.Instance().Delete(e2.Menu)
	} else {
		cache.Instance().Delete(e2.Menu + gconv.String(user.ID))
	}
}

// 用户注销
func SignOut(c *gin.Context) error {
	user := GetProfile(c)
	if user != nil {
		ClearMenuCache(user)
	}
	SessionList.Delete(user.ID)
	err := dao.NewAdminOnlineDaoImpl().Delete(gconv.String(user.ID))
	if err != nil {
		return err
	}
	err = session.Del(c, e2.Auth)
	if err != nil {
		return err
	}
	err = session.Del(c, e2.AdminInfo)
	if err != nil {
		return err
	}
	return nil
}

// 检查登陆名是否存在,存在返回true,否则false
func CheckLoginName(loginName string) bool {
	if admin, err := dao.NewAdminDaoImpl().FindByName(loginName); err != nil || admin.ID < 1 {
		return false
	} else {
		return true
	}
}

// 获得用户信息详情
func GetProfile(c *gin.Context) *model.Admin {
	tmp := session.Get(c, e2.AdminInfo)
	if tmp == nil {
		return nil
	}
	u := new(model.Admin)
	if s, ok := tmp.(string); ok {
		err := json.Unmarshal([]byte(s), &u)
		if err != nil {
			return nil
		}
	}
	return u
}

func GetMenuAuth() []model.Auth {
	filters := make([]interface{}, 0)
	filters = append(filters, "power_type < ?", 2)
	result, _ := dao.NewAuthDaoImpl().Find(1, 5000, filters...)
	return result
}

func GetAuth(user *model.Admin) []model.Auth {
	filters := make([]interface{}, 0)
	filters = append(filters, "status = ?", 1)
	if user.Level != 99 {
		//普通管理员
		adminAuthIds, _ := dao.NewRoleAuthDaoImpl().GetByRoleIds(user.RoleIds)
		adminAuthIdArr := strings.Split(adminAuthIds, ",")
		filters = append(filters, "id in (?)", adminAuthIdArr)
	}
	result, _ := dao.NewAuthDaoImpl().Find(1, 5000, filters...)
	return result
}

// 更新用户头像
func UpdateAvatarService(avatar string, c *gin.Context) error {
	user := GetProfile(c)
	if strings.HasPrefix(avatar, "/") == false {
		avatar = "/" + avatar
	}
	err := dao.NewAdminDaoImpl().Update(*user, map[string]interface{}{"avatar": avatar})
	if err != nil {
		return err
	}
	user.Avatar = avatar
	SaveUserToSession(*user, c)
	return nil
}

func ProfileEditService(f request.ProfileForm, c *gin.Context) error {
	user := GetProfile(c)
	err := dao.NewAdminDaoImpl().Update(*user, map[string]interface{}{
		"real_name": f.RealName,
		"email":     f.Email,
		"phone":     f.Phone,
		"remark":    f.Remark,
	})
	if err != nil {
		return err
	}
	user.Phone = f.Phone
	user.Email = f.Email
	user.RealName = f.RealName
	user.Remark = f.Remark
	SaveUserToSession(*user, c)
	return nil
}

func GetImgSavePath(path string) string {
	return strings.ReplaceAll(path, config.Instance().App.ImgUrlPath, config.Instance().App.ImgSavePath)
}

func GetLoginInfo(c *gin.Context) ([]model.LoginInfo, error) {
	user := GetProfile(c)
	filters := make([]interface{}, 0)
	filters = append(filters, "login_name = ?", user.LoginName)
	info, _, err := dao.NewLoginInfoImpl().FindByPage(1, 5, filters...)
	if err != nil {
		return nil, err
	}
	for k,_ := range info{
		info[k].IpAddr = "*.*.*.*"
	}
	return info, nil
}

//修改用户密码
func PwdEditHandlerService(profile request.PasswordForm, c *gin.Context) error {
	if profile.NewPwd == profile.OldPwd {
		return errors.New("新旧密码不能相同")
	}
	if profile.ConfirmPwd != profile.NewPwd {
		return errors.New("确认密码不一致")
	}
	//校验密码
	userCache := GetProfile(c)
	user, err := dao.NewAdminDaoImpl().FindByName(userCache.LoginName)
	if err != nil {
		return err
	}
	if user.Password != pkg.Md5([]byte(profile.OldPwd+user.Salt)) {
		return errors.New("原密码不正确")
	}

	//新校验密码
	pwd, salt := pkg.SetPassword(e2.DefaultSaltLen, profile.NewPwd)
	err = dao.NewAdminDaoImpl().Update(user, map[string]interface{}{
		"password": pwd,
		"salt":     salt,
	})
	if err != nil {
		return errors.New("保存数据失败")
	}
	// 密码和盐不存session 此处不用更新session缓存
	return nil
}
