package handlers

import (
	"github.com/porter-dev/porter-agent/pkg/redis"
	"github.com/spf13/viper"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	redisHost    string
	redisPort    string
	maxTailLines int64

	redisClient *redis.Client

	httpLogger = ctrl.Log.WithName("HTTP Server")
)

func init() {
	viper.SetDefault("REDIS_HOST", "porter-redis-master")
	viper.SetDefault("REDIS_PORT", "6379")
	maxTailLines = viper.GetInt64("MAX_TAIL_LINES")

	redisHost = viper.GetString("REDIS_HOST")
	redisPort = viper.GetString("REDIS_PORT")
	redisClient = redis.NewClient(redisHost, redisPort, "", "", redis.PODSTORE, maxTailLines)
}
