package main

import (
	"github.com/skr1ms/SkyPostDelivery/locker-agent/config"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/app"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log := logger.New("error")
		log.Fatal("Config error", err, nil)
	}

	app.Run(cfg)
}
