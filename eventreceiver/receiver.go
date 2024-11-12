package eventreceiver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"time"
)

type eventReceiver struct {
	cfg *Config

	nextLogs consumer.Logs

	server *http.Server

	obsrep   *receiverhelper.ObsReport
	settings *receiver.Settings
	logger   *zap.Logger
	tracer   trace.Tracer
}

type event struct {
	IPAddress string    `json:"ip_addr"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
}

func (e *event) ToOtel() plog.Logs {
	data := plog.NewLogs()
	rl := data.ResourceLogs().AppendEmpty()
	ss := rl.ScopeLogs().AppendEmpty()
	lr := ss.LogRecords().AppendEmpty()

	lr.Body().SetStr(e.Message)
	lr.SetTimestamp(pcommon.NewTimestampFromTime(e.Timestamp))
	lr.Attributes().PutStr("ip_address", e.IPAddress)
	lr.Attributes().PutStr("event_type", e.EventType)
	lr.Attributes().PutStr("user_id", e.UserID)

	return data
}

func (er *eventReceiver) Start(ctx context.Context, host component.Host) error {
	var (
		ln  net.Listener
		err error
	)
	ln, err = er.cfg.HTTP.ToListener(ctx)
	if err != nil {
		return fmt.Errorf("can't init http server: %s", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/report", er.Report).Methods(http.MethodPost)
	er.logger.Info("starting server",
		zap.Any("host", host),
		zap.Any("telemetry", er.settings.TelemetrySettings),
		zap.Any("router", r),
	)
	er.server, err = er.cfg.HTTP.ToServer(ctx, host, er.settings.TelemetrySettings, r)
	go func() {
		if errHTTP := er.server.Serve(ln); !errors.Is(errHTTP, http.ErrServerClosed) && errHTTP != nil {
			componentstatus.ReportStatus(host, componentstatus.NewFatalErrorEvent(errHTTP))
		}
	}()

	er.logger.Info("started http listener",
		zap.String("address", ln.Addr().String()),
	)
	return nil
}

func (er *eventReceiver) Shutdown(ctx context.Context) error {
	err := er.server.Shutdown(ctx)
	if err != nil {
		er.logger.Error("failed to stop http server", zap.Error(err))
	}
	return err
}

func newEventReceiver(cfg *Config, set *receiver.Settings, nextLogs consumer.Logs) (*eventReceiver, error) {
	var err error

	r := eventReceiver{
		cfg:      cfg,
		nextLogs: nextLogs,
		server:   nil,
		logger:   set.Logger,
		tracer:   set.TracerProvider.Tracer(typeStr),
		settings: set,
	}

	r.obsrep, err = receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             set.ID,
		Transport:              "http",
		ReceiverCreateSettings: *set,
	})

	if err != nil {
		return nil, fmt.Errorf("can't init telemetry: %s", err)
	}
	return &r, nil
}

func (er *eventReceiver) writeResponse(w http.ResponseWriter, err error) {
	switch err != nil {
	case true:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Content-Type", "application/json")
}

func (er *eventReceiver) Report(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)
	ctx := client.NewContext(r.Context(), client.Info{})

	ctx = er.obsrep.StartLogsOp(ctx)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		er.writeResponse(w, err)
		return
	}

	ev := event{}
	err = json.Unmarshal(bodyBytes, &ev)
	er.logger.Info("event received", zap.Any("event", ev))
	if err != nil {
		er.writeResponse(w, err)
		return
	}

	err = er.nextLogs.ConsumeLogs(ctx, ev.ToOtel())
	er.writeResponse(w, err)

	er.obsrep.EndLogsOp(ctx, "http", 1, err)
}
