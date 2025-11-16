package main

import (
	"log"

	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
	_ "github.com/skr1ms/SkyPostDelivery/drone-service/docs"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/app"
)

// @title           SkyPost Drone Service API
// @version         1.0
// @description     API для управления дронами доставки
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  skr1ms13666@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8081
// @BasePath  /api/v1

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
