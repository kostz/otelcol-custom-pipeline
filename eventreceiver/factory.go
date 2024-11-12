package eventreceiver

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr  = "eventreceiver"
	httpPort = "5510"
)

func createDefaultConfig() component.Config {
	return &Config{
		HTTP: &confighttp.ServerConfig{
			Endpoint: fmt.Sprintf("0.0.0.0:%s", httpPort),
		},
	}
}

func createEventReceiver(ctx context.Context, set receiver.Settings, config component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	cfg := config.(*Config)
	return newEventReceiver(cfg, &set, nextConsumer)
}

// NewFactory creates a receiver factory
func NewFactory() receiver.Factory {
	cfgType, _ := component.NewType(typeStr)
	return receiver.NewFactory(
		cfgType,
		createDefaultConfig,
		receiver.WithLogs(createEventReceiver, component.StabilityLevelDevelopment),
	)
}
