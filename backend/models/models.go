package models

import "gorm.io/gorm"

type SensorType struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	DataType    string `gorm:"not null" json:"dataType"`
	RegCount    int    `gorm:"not null" json:"regCount"`
	Color       string `json:"color"`
}

type GatewayTemplate struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"unique;not null" json:"name"`
	Model       string `json:"model"`
	Description string `json:"description"`
}

type RegisterBinding struct {
	ID                uint   `gorm:"primaryKey" json:"id"`
	GatewayTemplateID uint   `gorm:"not null;index" json:"gatewayTemplateId"`
	SensorTypeID      uint   `gorm:"not null" json:"sensorTypeId"`
	StartAddress      int    `gorm:"not null" json:"startAddress"`
	Label             string `json:"label"`
	SensorType        SensorType `gorm:"foreignKey:SensorTypeID" json:"sensorType"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&SensorType{}, &GatewayTemplate{}, &RegisterBinding{})
}

func SeedData(db *gorm.DB) error {
	sensors := []SensorType{
		{Name: "16位温度传感器", Description: "PT100 温度采集，16位有符号整数", DataType: "int16", RegCount: 1, Color: "#ef4444"},
		{Name: "32位浮点压力传感器", Description: "工业压力变送器，IEEE 754 单精度浮点", DataType: "float32", RegCount: 2, Color: "#3b82f6"},
		{Name: "32位浮点流量传感器", Description: "电磁流量计，累计流量", DataType: "float32", RegCount: 2, Color: "#10b981"},
		{Name: "16位湿度传感器", Description: "环境湿度采集，0-100%", DataType: "uint16", RegCount: 1, Color: "#8b5cf6"},
		{Name: "64位双精度电能表", Description: "三相电能计量，累计用电量", DataType: "float64", RegCount: 4, Color: "#f59e0b"},
		{Name: "16位开关状态", Description: "DI 数字输入，8路开关量", DataType: "uint16", RegCount: 1, Color: "#6b7280"},
		{Name: "32位累计脉冲", Description: "脉冲计数，无符号32位", DataType: "uint32", RegCount: 2, Color: "#ec4899"},
		{Name: "16位模拟输出", Description: "AO 模拟输出，4-20mA", DataType: "uint16", RegCount: 1, Color: "#14b8a6"},
	}

	for _, s := range sensors {
		var count int64
		db.Model(&SensorType{}).Where("name = ?", s.Name).Count(&count)
		if count == 0 {
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		}
	}

	templates := []GatewayTemplate{
		{Name: "EGW-200 标准版", Model: "EGW-200", Description: "标准工业级边缘网关，支持 100 个保持寄存器"},
		{Name: "EGW-400 增强版", Model: "EGW-400", Description: "增强型边缘网关，双网口冗余"},
	}

	for _, t := range templates {
		var count int64
		db.Model(&GatewayTemplate{}).Where("name = ?", t.Name).Count(&count)
		if count == 0 {
			if err := db.Create(&t).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
