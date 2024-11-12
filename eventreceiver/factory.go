package eventreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr = "eventreceiver"
)

func createDefaultConfig() component.Config {
	return Config{}
}

func createEventReceiver(ctx context.Context, set receiver.Settings, config component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	cfg := config.(*Config)
	return newEventReceiver(cfg, &set, nextConsumer)
}

func NewFactory() receiver.Factory {
	cfgType, _ := component.NewType(typeStr)
	return receiver.NewFactory(
		cfgType,
		createDefaultConfig,
		receiver.WithLogs(createEventReceiver, component.StabilityLevelDevelopment),
	)
}
