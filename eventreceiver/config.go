package eventreceiver

import "go.opentelemetry.io/collector/config/confighttp"

// Config holds the receiver server configuration
type Config struct {
	HTTP *confighttp.ServerConfig `mapstructure:"http"`
}
