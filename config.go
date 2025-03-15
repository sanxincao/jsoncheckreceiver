package jsoncheckreceiver

import (
	"time"

	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	ServerURL                      string   `mapstructure:"endpoint"`
	Targets                        []Target `mapstructure:"targets"`
}

type Target struct {
	Target string `mapstructure:"target"`
}
