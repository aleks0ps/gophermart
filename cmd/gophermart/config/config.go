package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	defaultRunAddress     = "localhost:8088"
	defaultDatabaseURI    = "postgres://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable"
	defaultAccrualAddress = "localhost:8080"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ParseOptions() *Config {
	opts := Config{
		RunAddress:     defaultRunAddress,
		DatabaseURI:    defaultDatabaseURI,
		AccrualAddress: defaultAccrualAddress,
	}
	if err := env.Parse(&opts); err != nil {
		fmt.Println("failed:", err)
	}
	flag.StringVar(&opts.RunAddress, "a", opts.RunAddress, "Listen address:port")
	flag.StringVar(&opts.DatabaseURI, "d", opts.DatabaseURI, "Postgres connection string")
	flag.StringVar(&opts.AccrualAddress, "r", opts.AccrualAddress, "Accrual system address")
	flag.Parse()
	return &opts
}
