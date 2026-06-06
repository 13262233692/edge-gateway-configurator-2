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

func buildOccupiedMap(existing []models.RegisterBinding, excludeID uint) (map[int]models.RegisterBinding, error) {
	occupied := make(map[int]models.RegisterBinding)
	for _, b := range existing {
		if b.ID == excludeID {
			continue
		}
		var sensor models.SensorType
		if err := database.DB.First(&sensor, b.SensorTypeID).Error; err != nil {
			return nil, fmt.Errorf("传感器类型查询失败: %w", err)
		}
		regCount := sensor.RegCount
		if regCount <= 0 {
			regCount = 1
		}
		b.SensorType = sensor
		for i := 0; i < regCount; i++ {
			addr := b.StartAddress + i
			occupied[addr] = b
		}
	}
	return occupied, nil
}

func checkAddressConflict(startAddress int, regCount int, occupied map[int]models.RegisterBinding) (bool, string) {
	if regCount <= 0 {
		return true, "传感器寄存器数量无效"
	}
	endAddress := startAddress + regCount - 1
	if endAddress > 40100 {
		return true, fmt.Sprintf("地址范围超出边界: %d-%d 超过 40100", startAddress, endAddress)
	}
	for i := 0; i < regCount; i++ {
		addr := startAddress + i
		if conflict, exists := occupied[addr]; exists {
			conflictEnd := conflict.StartAddress + conflict.SensorType.RegCount - 1
			return true, fmt.Sprintf(
				"地址 %d 与传感器 \"%s\" (占用 %d-%d) 重叠",
				addr, conflict.SensorType.Name, conflict.StartAddress, conflictEnd,
			)
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

	if sensor.RegCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "传感器寄存器数量配置无效"})
		return
	}

	var existing []models.RegisterBinding
	database.DB.Where("gateway_template_id = ?", req.GatewayTemplateID).Find(&existing)

	occupied, err := buildOccupiedMap(existing, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if conflict, msg := checkAddressConflict(req.StartAddress, sensor.RegCount, occupied); conflict {
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

	if sensor.RegCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "传感器寄存器数量配置无效"})
		return
	}

	var existing []models.RegisterBinding
	database.DB.Where("gateway_template_id = ?", req.GatewayTemplateID).Find(&existing)

	occupied, err := buildOccupiedMap(existing, binding.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if conflict, msg := checkAddressConflict(req.StartAddress, sensor.RegCount, occupied); conflict {
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
			if s.RegCount <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("传感器 \"%s\" 寄存器数量配置无效", s.Name)})
				return
			}
			sensorMap[b.SensorTypeID] = s
		}
	}

	occupied := make(map[int]string)
	for i, b := range req.Bindings {
		s := sensorMap[b.SensorTypeID]
		end := b.StartAddress + s.RegCount - 1
		if end > 40100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个传感器 \"%s\" 地址范围 %d-%d 超出 40100 边界", i+1, s.Name, b.StartAddress, end)})
			return
		}
		for j := 0; j < s.RegCount; j++ {
			addr := b.StartAddress + j
			if existName, exists := occupied[addr]; exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("地址 %d 冲突: \"%s\" 与 \"%s\" 重叠", addr, existName, s.Name)})
				return
			}
			occupied[addr] = s.Name
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
