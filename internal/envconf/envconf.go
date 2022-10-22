package envconf

import (
	"github.com/porter-dev/porter-agent/pkg/httpclient"
	"github.com/porter-dev/porter/api/server/shared/config/env"
)

type LogStoreConf struct {
	LogStoreAddress     string `env:"LOG_STORE_ADDRESS,default=:9096"`
	LogStoreHTTPAddress string `env:"LOG_STORE_HTTP_ADDRESS"`
	LogStoreKind        string `env:"LOG_STORE_KIND,default=memory"`
}
type EnvDecoderConf struct {
	Debug      bool   `env:"DEBUG,default=true"`
	SentryDSN  string `env:"SENTRY_DSN"`
	SentryEnv  string `env:"SENTRY_ENV,default=dev"`
	ServerPort uint   `env:"SERVER_PORT,default=10001"`

	LogStoreConf   LogStoreConf
	HTTPClientConf httpclient.HTTPClientConf
	DBConf         env.DBConf
}
