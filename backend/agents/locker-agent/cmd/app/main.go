package main

import (
	"log"

	"github.com/skr1ms/SkyPostDelivery/locker-agent/config"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/app"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
