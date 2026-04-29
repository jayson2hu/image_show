package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/model"
)

type packageRequest struct {
	Name      string  `json:"name" binding:"required"`
	Credits   float64 `json:"credits" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	ValidDays int     `json:"valid_days" binding:"required"`
	SortOrder int     `json:"sort_order"`
	Status    int     `json:"status"`
}

func Packages(c *gin.Context) {
	var items []model.Package
	if err := model.DB.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list packages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func AdminPackages(c *gin.Context) {
	var items []model.Package
	if err := model.DB.Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list packages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func AdminCreatePackage(c *gin.Context) {
	var req packageRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Credits <= 0 || req.Price <= 0 || req.ValidDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	pkg := packageFromRequest(req)
	if err := model.DB.Create(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create package"})
		return
	}
	c.JSON(http.StatusOK, pkg)
}

func AdminUpdatePackage(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req packageRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Credits <= 0 || req.Price <= 0 || req.ValidDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := model.DB.Model(&model.Package{}).Where("id = ?", id).Updates(packageFromRequest(req)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update package"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func AdminDeletePackage(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := model.DB.Delete(&model.Package{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete package"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func packageFromRequest(req packageRequest) model.Package {
	status := req.Status
	if status == 0 {
		status = 1
	}
	return model.Package{
		Name:      req.Name,
		Credits:   req.Credits,
		Price:     req.Price,
		ValidDays: req.ValidDays,
		SortOrder: req.SortOrder,
		Status:    status,
	}
}
