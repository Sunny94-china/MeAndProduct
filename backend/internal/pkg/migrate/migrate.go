package migrate

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"meandproduct/internal/config"
)

// Run executes migrations.
//   up=true, downN=0 → Up()
//   up=false, downN>0 → Steps(-downN)
func Run(cfg *config.Config, up bool, downN int) error {
	if up && downN > 0 {
		return errors.New("specify either up or downN, not both")
	}
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&multiStatements=true&parseTime=true&loc=Local",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.DB,
	)

	m, err := migrate.New(
		cfg.Migration.Path,
		"mysql://"+dsn,
	)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if up {
		slog.Info("running migrations UP", "path", cfg.Migration.Path)
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				slog.Info("migrations: no change")
				return nil
			}
			return fmt.Errorf("migrate up: %w", err)
		}
		slog.Info("migrations up OK")
		return nil
	}
	if downN <= 0 {
		return nil
	}
	slog.Info("running migrations DOWN", "steps", downN)
	if err := m.Steps(-downN); err != nil {
		return fmt.Errorf("migrate down steps %d: %w", downN, err)
	}
	slog.Info("migrations down OK")
	return nil
}
