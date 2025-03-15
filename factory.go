package jsoncheckreceiver

import (
	"context"
	"errors"

	"github.com/sanxincao/icmpcheckreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	cfg := scraperhelper.NewDefaultControllerConfig()

	return &Config{
		ControllerConfig: cfg,
		ServerURL:        "http://127.0.0.1", // 默认的服务器URL
		Targets:          []Target{},
	}
}

func createMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	receiverCfg, ok := cfg.(*Config)
	if !ok {
		return nil, errors.New("config is not a valid jsoncheck receiver config")
	}

	opts := []scraperhelper.ScraperControllerOption{}

	for _, target := range receiverCfg.Targets {
		jsonScraper, err := newScraper(set.Logger, receiverCfg.ServerURL, target.Target)
		if err != nil {
			return nil, err
		}

		scraper, err := scraperhelper.NewScraper(metadata.Type, jsonScraper.Scrape)
		if err != nil {
			return nil, err
		}

		opts = append(opts, scraperhelper.AddScraper(scraper))
	}

	return scraperhelper.NewScraperControllerReceiver(&receiverCfg.ControllerConfig, set, nextConsumer, opts...)
}
