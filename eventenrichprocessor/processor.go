package eventenrichprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// EventEnrichProcessor keeps the component of a processor
type EventEnrichProcessor struct {
	ipResolveClient IPResolveClient
	logger          *zap.Logger
	tracer          trace.Tracer
}

// IPMetadata keeps the dataset to be added to log records
type IPMetadata struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
}

const (
	// IPAddressKey defines the key name with ip address
	IPAddressKey = "ip_address"
)

// NewEventEnrichProcessor creates the instance of Event enrich processor
func NewEventEnrichProcessor(ipResolveClient IPResolveClient, set *processor.Settings) (*EventEnrichProcessor, error) {
	return &EventEnrichProcessor{
		ipResolveClient: ipResolveClient,
		logger:          set.Logger,
		tracer:          set.TracerProvider.Tracer(typeStr),
	}, nil
}

// IPResolveClient defines the interface for IP Resolve service
type IPResolveClient interface {
	ResolveIP(context.Context, string) ([]attribute.KeyValue, error)
}

type ipResolveClientImpl struct {
	logger *zap.Logger
	url    string
}

func (i ipResolveClientImpl) ResolveIP(ctx context.Context, ipAddress string) ([]attribute.KeyValue, error) {
	res := []attribute.KeyValue{}
	data := IPMetadata{}

	rs, err := http.Get(fmt.Sprintf("%s/%s", i.url, ipAddress))
	if err != nil {
		i.logger.Error("can't call ip service",
			zap.Int("ret_code", rs.StatusCode),
		)
		return res, nil
	}
	rsData, err := io.ReadAll(rs.Body)
	if err != nil {
		i.logger.Error("can't read response ",
			zap.Error(err),
		)
		return res, nil
	}

	err = json.Unmarshal(rsData, &data)
	if err != nil {
		i.logger.Error("can't unmarshal response ",
			zap.Error(err),
		)
		return res, nil
	}

	res = append(res, []attribute.KeyValue{
		attribute.String("city", data.City),
		attribute.String("country", data.Country),
		attribute.String("currency", data.Currency),
		attribute.String("region", data.Region),
		attribute.String("timezone", data.Timezone),
	}...)

	i.logger.Info("response ",
		zap.Any("response", res),
	)

	return res, nil
}

// NewIPResolveClient creates the client to IP resolve service
func NewIPResolveClient(url string, logger *zap.Logger) (IPResolveClient, error) {
	return &ipResolveClientImpl{
		url:    url,
		logger: logger,
	}, nil
}

// ProcessLogs enriches log recoeds with metadata by ip address
func (p *EventEnrichProcessor) ProcessLogs(
	ctx context.Context,
	logs plog.Logs,
) (plog.Logs, error) {
	ctx, span := p.tracer.Start(ctx, "enrich")
	cnt := 0
	defer span.End()

	logs.ResourceLogs().RemoveIf(func(resourceLogs plog.ResourceLogs) bool {
		resourceLogs.ScopeLogs().RemoveIf(func(scopeLogs plog.ScopeLogs) bool {
			scopeLogs.LogRecords().RemoveIf(func(log plog.LogRecord) bool {

				ipAddress, ok := log.Attributes().Get(IPAddressKey)
				if !ok {
					p.logger.Warn("no ip address provided in log message")
				}

				attrs, err := p.ipResolveClient.ResolveIP(ctx, ipAddress.AsString())
				p.logger.Info("ip resolved",
					zap.String("ip", ipAddress.AsString()),
					zap.Any("attrs", attrs),
					zap.Error(err),
				)
				if err != nil {
					p.logger.Warn("can't resolve ip", zap.Error(err))
				}
				for _, a := range attrs {
					log.Attributes().PutStr(string(a.Key), a.Value.AsString())
				}
				cnt++
				return false
			})
			return false
		})
		return false
	})
	span.SetAttributes(attribute.Int("processed", cnt))

	return logs, nil
}
