package main

import (
	"edge-gateway-configurator/database"
	"edge-gateway-configurator/handlers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		sensors := api.Group("/sensors")
		{
			sensors.GET("", handlers.GetSensorTypes)
			sensors.POST("", handlers.CreateSensorType)
			sensors.PUT("/:id", handlers.UpdateSensorType)
			sensors.DELETE("/:id", handlers.DeleteSensorType)
		}

		templates := api.Group("/templates")
		{
			templates.GET("", handlers.GetGatewayTemplates)
			templates.POST("", handlers.CreateGatewayTemplate)
			templates.PUT("/:id", handlers.UpdateGatewayTemplate)
			templates.DELETE("/:id", handlers.DeleteGatewayTemplate)
		}

		bindings := api.Group("/bindings")
		{
			bindings.GET("", handlers.GetRegisterBindings)
			bindings.POST("", handlers.CreateRegisterBinding)
			bindings.PUT("/:id", handlers.UpdateRegisterBinding)
			bindings.DELETE("/:id", handlers.DeleteRegisterBinding)
			bindings.POST("/save-all", handlers.SaveAllBindings)
		}
	}

	log.Println("服务器启动于 http://localhost:8080")
	r.Run(":8080")
}
