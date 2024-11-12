package eventreceiver

import "go.opentelemetry.io/collector/config/confighttp"

type Config struct {
	HTTP *confighttp.ServerConfig `mapstructure:"http"`
}
