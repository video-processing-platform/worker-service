package main

import (
	"fmt"
	"github.com/alimarzban99/video-processor-service/config"
	"github.com/alimarzban99/video-processor-service/pkg/cache"
	"github.com/alimarzban99/video-processor-service/pkg/database"
	"github.com/alimarzban99/video-processor-service/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {

	config.Load()

	if err := database.Init(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := cache.Init(); err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	rabbitmq.Init()

	fmt.Println("sdfsdf")

	appConfig := config.Cfg.App
	gin.SetMode(appConfig.Environment)
	router := gin.New()

	runPort := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	err := router.Run(runPort)
	if err != nil {
		log.Fatal(err)
	}

}
