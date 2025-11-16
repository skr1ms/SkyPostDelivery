package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/migrator"
)

func runMigrations(db *pgxpool.Pool) error {
	return migrator.Run(
		db.Config().ConnConfig.ConnString(),
		"migrations",
	)
}
