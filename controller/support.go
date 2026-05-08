package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

func SupportContact(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"credit_exhausted_message":           model.GetSettingValue("credit_exhausted_message", "额度已用完，可以注册账号获取新用户积分；如需人工开通或咨询套餐，请联系管理员。"),
		"credit_exhausted_wechat_qrcode_url": model.GetSettingValue("credit_exhausted_wechat_qrcode_url", ""),
		"credit_exhausted_qq":                model.GetSettingValue("credit_exhausted_qq", ""),
		"manual_recharge_enabled":            model.GetSettingValue("manual_recharge_enabled", "true"),
		"manual_recharge_wechat_id":          model.GetSettingValue("manual_recharge_wechat_id", ""),
		"manual_recharge_wechat_qrcode_url":  model.GetSettingValue("manual_recharge_wechat_qrcode_url", ""),
		"manual_recharge_qq":                 model.GetSettingValue("manual_recharge_qq", ""),
		"manual_recharge_note":               model.GetSettingValue("manual_recharge_note", "请联系管理员人工充值，并备注账号邮箱和需要开通的套餐。"),
	})
}
