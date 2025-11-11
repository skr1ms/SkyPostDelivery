package main

import (
	"log"

	"github.com/skr1ms/hitech-ekb/config"
	_ "github.com/skr1ms/hitech-ekb/docs"
	"github.com/skr1ms/hitech-ekb/internal/app"
)

// @title           hiTech Drone Delivery API
// @version         1.0
// @description     API для системы доставки посылок дронами
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@hitech.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer <token>

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
