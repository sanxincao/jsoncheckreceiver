package metadata

import (
	"go.opentelemetry.io/collector/component"
)

var (
	Type = component.MustNewType("jsoncheck")
)

const (
	MetricsStability = component.StabilityLevelDevelopment
)
