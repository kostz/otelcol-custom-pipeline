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

type EventEnrichProcessor struct {
	ipResolveClient IPResolveClient
	logger          *zap.Logger
	tracer          trace.Tracer
}

type IPMetadata struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
}

const (
	IpAddressKey = "ip_address"
)

func NewEventEnrichProcessor(ipResolveClient IPResolveClient, set *processor.Settings) (*EventEnrichProcessor, error) {
	return &EventEnrichProcessor{
		ipResolveClient: ipResolveClient,
		logger:          set.Logger,
		tracer:          set.TracerProvider.Tracer(typeStr),
	}, nil
}

type IPResolveClient interface {
	ResolveIP(context.Context, string) ([]attribute.KeyValue, error)
}

type ipResolveClientImpl struct {
	url string
}

func (i ipResolveClientImpl) ResolveIP(ctx context.Context, ipAddress string) ([]attribute.KeyValue, error) {
	res := []attribute.KeyValue{}
	data := IPMetadata{}

	rs, err := http.Get(fmt.Sprintf("%s/%s", i.url, ipAddress))
	if err != nil {
		return res, nil
	}
	rsData, err := io.ReadAll(rs.Body)
	if err != nil {
		return res, nil
	}

	err = json.Unmarshal(rsData, &data)
	if err != nil {
		return res, nil
	}

	res = append(res, []attribute.KeyValue{
		attribute.String("city", data.City),
		attribute.String("country", data.Country),
		attribute.String("currency", data.Currency),
		attribute.String("region", data.Region),
		attribute.String("timezone", data.Timezone),
	}...)

	return res, nil
}

func NewIPResolveClient(url string) (IPResolveClient, error) {
	return &ipResolveClientImpl{
		url: url,
	}, nil
}

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
				ipAddress, _ := log.Attributes().Get(IpAddressKey)
				attrs, err := p.ipResolveClient.ResolveIP(ctx, ipAddress.Str())
				if err != nil {
					return false
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
