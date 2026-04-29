package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/service"
	"gorm.io/gorm"
)

type createOrderRequest struct {
	PackageID int64  `json:"package_id" binding:"required"`
	PayMethod string `json:"pay_method" binding:"required"`
}

func CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PackageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := service.CreatePaymentOrder(c.GetInt64("userID"), req.PackageID, req.PayMethod)
	if err != nil {
		writePaymentError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func GetOrder(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	order, err := service.GetUserOrder(c.GetInt64("userID"), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func PaymentNotify(c *gin.Context) {
	params, err := paymentNotifyParams(c)
	if err != nil || len(params) == 0 {
		c.String(http.StatusOK, "fail")
		return
	}
	result, err := service.HandlePaymentNotify(params)
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	if result.Status != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

func paymentNotifyParams(c *gin.Context) (map[string]string, error) {
	params := map[string]string{}
	if c.Request.Method == http.MethodPost {
		if err := c.Request.ParseForm(); err != nil {
			return nil, err
		}
		for key := range c.Request.PostForm {
			params[key] = c.Request.PostForm.Get(key)
		}
		return params, nil
	}
	for key := range c.Request.URL.Query() {
		params[key] = c.Request.URL.Query().Get(key)
	}
	return params, nil
}

func writePaymentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPayMethod), errors.Is(err, service.ErrPackageUnavailable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrPaymentNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment not configured"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
	}
}
