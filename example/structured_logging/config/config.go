package config

import (
	"context"

	"github.com/m0rjc/goconfig"
)

type DatabaseConfig struct {
	Host     string `key:"DB_HOST" required:"true"`
	Port     int    `key:"DB_PORT" required:"true" min:"1024" max:"65535"`
	Username string `key:"DB_USER" required:"true"`
	Password string `key:"DB_PASS" required:"true" pattern:"^.{8,}$"` // min 8 chars
	Database string `key:"DB_NAME" required:"true"`
	Timeout  int    `key:"DB_TIMEOUT" default:"30" min:"1" max:"300"`
}

func Load(ctx context.Context) (*DatabaseConfig, error) {
	cfg := &DatabaseConfig{}
	err := goconfig.Load(ctx, cfg)
	return cfg, err
}
