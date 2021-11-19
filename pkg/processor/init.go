package processor

import "github.com/spf13/viper"

var (
	redisHost    string
	redisPort    string
	maxTailLines int64
)

func init() {
	viper.SetDefault("REDIS_HOST", "porter-redis-master")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("MAX_TAIL_LINES", int64(100))
	viper.AutomaticEnv()

	redisHost = viper.GetString("REDIS_HOST")
	redisPort = viper.GetString("REDIS_PORT")
	maxTailLines = viper.GetInt64("MAX_TAIL_LINES")
}
