package operations

import (
	"runtime"

	"github.com/spf13/viper"
)

func SetupConfig() {
	viper.SetEnvPrefix("dyntcp")
	viper.AutomaticEnv()

	viper.SetDefault("consul_addr", "http://127.0.0.1:8500")
	viper.SetDefault("gomaxprocs", runtime.NumCPU())
	viper.SetDefault("ports_min", 1024)
	viper.SetDefault("ports_max", 49151)
	viper.SetDefault("peek_listener", "")
	viper.SetDefault(
		"allowed_listeners",
		map[string]string{"local": "127.0.0.1", "any": "0.0.0.0"},
	)
}
