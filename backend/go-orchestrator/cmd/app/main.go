package main

import (
	"log"

	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	_ "github.com/skr1ms/SkyPostDelivery/go-orchestrator/docs"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/app"
)

// @title           SkyPost Delivery API
// @version         1.0
// @description     API для системы доставки посылок
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  skr1ms13666@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
