package handlers

import (
	"edge-gateway-configurator/database"
	"edge-gateway-configurator/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BindingRequest struct {
	GatewayTemplateID uint   `json:"gatewayTemplateId" binding:"required"`
	SensorTypeID      uint   `json:"sensorTypeId" binding:"required"`
	StartAddress      int    `json:"startAddress" binding:"required,min=40001,max=40100"`
	Label             string `json:"label"`
}

type SaveAllRequest struct {
	GatewayTemplateID uint              `json:"gatewayTemplateId" binding:"required"`
	Bindings          []BindingRequest `json:"bindings"`
}

func checkOverlap(existing []models.RegisterBinding, newSensorID uint, newStart, newCount int, excludeID uint) (bool, string) {
	newEnd := newStart + newCount - 1
	if newEnd > 40100 {
		return true, "传感器地址范围超出 40100 边界"
	}

	for _, b := range existing {
		if b.ID == excludeID {
			continue
		}
		existEnd := b.StartAddress + b.SensorType.RegCount - 1
		if !(newEnd < b.StartAddress || newStart > existEnd) {
			return true, fmt.Sprintf("与传感器 \"%s\" (地址 %d-%d) 内存重叠", b.SensorType.Name, b.StartAddress, existEnd)
		}
	}
	return false, ""
}

func GetSensorTypes(c *gin.Context) {
	var sensors []models.SensorType
	if err := database.DB.Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensors)
}

func CreateSensorType(c *gin.Context) {
	var sensor models.SensorType
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sensor)
}

func UpdateSensorType(c *gin.Context) {
	id := c.Param("id")
	var sensor models.SensorType
	if err := database.DB.First(&sensor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "传感器类型不存在"})
		return
	}
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Save(&sensor)
	c.JSON(http.StatusOK, sensor)
}

func DeleteSensorType(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.SensorType{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetGatewayTemplates(c *gin.Context) {
	var templates []models.GatewayTemplate
	if err := database.DB.Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func CreateGatewayTemplate(c *gin.Context) {
	var tpl models.GatewayTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&tpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tpl)
}

func UpdateGatewayTemplate(c *gin.Context) {
	id := c.Param("id")
	var tpl models.GatewayTemplate
	if err := database.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Save(&tpl)
	c.JSON(http.StatusOK, tpl)
}

func DeleteGatewayTemplate(c *gin.Context) {
	id := c.Param("id")
	database.DB.Where("gateway_template_id = ?", id).Delete(&models.RegisterBinding{})
	if err := database.DB.Delete(&models.GatewayTemplate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetRegisterBindings(c *gin.Context) {
	tplID := c.Query("gatewayTemplateId")
	var bindings []models.RegisterBinding
	query := database.DB.Preload("SensorType")
	if tplID != "" {
		query = query.Where("gateway_template_id = ?", tplID)
	}
	if err := query.Find(&bindings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bindings)
}

func CreateRegisterBinding(c *gin.Context) {
	var req BindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sensor models.SensorType
	if err := database.DB.First(&sensor, req.SensorTypeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "传感器类型不存在"})
		return
	}

	var existing []models.RegisterBinding
	database.DB.Preload("SensorType").Where("gateway_template_id = ?", req.GatewayTemplateID).Find(&existing)

	if overlap, msg := checkOverlap(existing, req.SensorTypeID, req.StartAddress, sensor.RegCount, 0); overlap {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址冲突: " + msg})
		return
	}

	binding := models.RegisterBinding{
		GatewayTemplateID: req.GatewayTemplateID,
		SensorTypeID:      req.SensorTypeID,
		StartAddress:      req.StartAddress,
		Label:             req.Label,
	}
	if err := database.DB.Create(&binding).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	database.DB.Preload("SensorType").First(&binding, binding.ID)
	c.JSON(http.StatusCreated, binding)
}

func UpdateRegisterBinding(c *gin.Context) {
	id := c.Param("id")
	var binding models.RegisterBinding
	if err := database.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "绑定不存在"})
		return
	}

	var req BindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sensor models.SensorType
	if err := database.DB.First(&sensor, req.SensorTypeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "传感器类型不存在"})
		return
	}

	var existing []models.RegisterBinding
	database.DB.Preload("SensorType").Where("gateway_template_id = ?", req.GatewayTemplateID).Find(&existing)

	if overlap, msg := checkOverlap(existing, req.SensorTypeID, req.StartAddress, sensor.RegCount, binding.ID); overlap {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址冲突: " + msg})
		return
	}

	binding.SensorTypeID = req.SensorTypeID
	binding.StartAddress = req.StartAddress
	binding.Label = req.Label
	database.DB.Save(&binding)
	database.DB.Preload("SensorType").First(&binding, binding.ID)
	c.JSON(http.StatusOK, binding)
}

func DeleteRegisterBinding(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.RegisterBinding{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func SaveAllBindings(c *gin.Context) {
	var req SaveAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensorMap := make(map[uint]models.SensorType)
	for _, b := range req.Bindings {
		if _, ok := sensorMap[b.SensorTypeID]; !ok {
			var s models.SensorType
			if err := database.DB.First(&s, b.SensorTypeID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "传感器类型不存在"})
				return
			}
			sensorMap[b.SensorTypeID] = s
		}
	}

	for i, b := range req.Bindings {
		s := sensorMap[b.SensorTypeID]
		end := b.StartAddress + s.RegCount - 1
		if end > 40100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个传感器地址范围超出 40100 边界", i+1)})
			return
		}
		for j := 0; j < i; j++ {
			prev := req.Bindings[j]
			prevS := sensorMap[prev.SensorTypeID]
			prevEnd := prev.StartAddress + prevS.RegCount - 1
			if !(end < prev.StartAddress || b.StartAddress > prevEnd) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "传感器 \"" + s.Name + "\" 与 \"" + prevS.Name + "\" 地址重叠"})
				return
			}
		}
	}

	database.DB.Where("gateway_template_id = ?", req.GatewayTemplateID).Delete(&models.RegisterBinding{})

	var created []models.RegisterBinding
	for _, b := range req.Bindings {
		binding := models.RegisterBinding{
			GatewayTemplateID: req.GatewayTemplateID,
			SensorTypeID:      b.SensorTypeID,
			StartAddress:      b.StartAddress,
			Label:             b.Label,
		}
		if err := database.DB.Create(&binding).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		database.DB.Preload("SensorType").First(&binding, binding.ID)
		created = append(created, binding)
	}

	c.JSON(http.StatusOK, created)
}
