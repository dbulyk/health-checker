package configs

import (
	"flag"
	"log/slog"

	"time"

	"github.com/caarlos0/env/v6"
)

type Checker struct {
	Interval  time.Duration `env:"CHECK_INTERVAL"`
	Address   string        `env:"ADDRESS"`
	Port      string        `env:"PORT"`
	DebugMode bool          `env:"DEBUG_MODE"`
}

var checker Checker

func GetCheckerCfg() Checker {
	flag.DurationVar(&checker.Interval, "i", 60*time.Second, "check interval")
	flag.StringVar(&checker.Address, "a", "localhost", "address")
	flag.StringVar(&checker.Port, "p", "8080", "port")
	flag.BoolVar(&checker.DebugMode, "d", false, "debug mode")
	flag.Parse()

	err := env.Parse(&checker)
	if err != nil {
		slog.Error("ошибка парсинга конфига: %v", err)
		panic(err)
	}
	return checker
}
