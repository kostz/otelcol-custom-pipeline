package eventenrichprocessor

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

const (
	typeStr = "eventenrichprocessor"
)

func createLogsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	localConfig, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unable to parse config")
	}

	ipResolveClient, err := NewIPResolveClient(localConfig.IPResolveServiceURL, set.Logger)
	if err != nil {
		zap.Error(err)
	}

	logsProc, err := NewEventEnrichProcessor(ipResolveClient, &set)
	if err != nil {
		zap.Error(err)
	}

	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		logsProc.ProcessLogs,
		processorhelper.WithCapabilities(
			consumer.Capabilities{
				MutatesData: true,
			},
		),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func NewFactory() processor.Factory {
	cfgType, _ := component.NewType(typeStr)
	return processor.NewFactory(
		cfgType,
		createDefaultConfig,
		processor.WithLogs(
			createLogsProcessor,
			component.StabilityLevelDevelopment,
		),
	)
}
