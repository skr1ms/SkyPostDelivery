package main

import (
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	_ "github.com/skr1ms/SkyPostDelivery/go-orchestrator/docs"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/app"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

// @title           SkyPost Delivery API
// @version         1.0
// @description     API for a package delivery system by drones
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  skr1ms13666@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	cfg, err := config.New()
	log := logger.New("error")
	if err != nil {
		log.Error("Config error", err, nil)
		return
	}

	app.Run(cfg)
}
