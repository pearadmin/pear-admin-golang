package captcha

import (
	"github.com/mojocn/base64Captcha"
	"image/color"
	e "pear-admin-go/app/global/e"
)

// 设置自带的store
var store = base64Captcha.DefaultMemStore

//生成验证码
func CaptMake() (id, b64s string, err error) {
	driver := mathCaptcha()
	captcha := base64Captcha.NewCaptcha(driver, store)
	lid, lb64s, lerr := captcha.Generate()
	return lid, lb64s, lerr
}

// 数字运算验证码
func mathCaptcha() base64Captcha.Driver {
	return base64Captcha.NewDriverMath(
		e.ImgHeight,
		e.ImgWidth,
		0,
		0,
		&color.RGBA{0, 0, 0, 0},
		[]string{"RitaSmith.ttf"},
	)
}

// 字符验证码
func stringCaptcha() base64Captcha.Driver {
	var driver base64Captcha.Driver
	var driverString base64Captcha.DriverString
	captchaConfig := base64Captcha.DriverString{
		Height:          e.ImgHeight,
		Width:           e.ImgWidth,
		NoiseCount:      0,     // 干扰字母
		ShowLineOptions: 1 | 3, // 干扰线
		Length:          e.ImgKeyLength,
		Source:          "qwertyuioplkjhgfdsazxcvbnm",
	}
	driverString = captchaConfig
	driver = driverString.ConvertFonts()
	return driver
}

//验证captcha是否正确
func CaptVerify(id string, capt string) bool {
	if store.Verify(id, capt, false) {
		return true
	} else {
		return false
	}
}
