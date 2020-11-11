package operations

import (
	"runtime"

	"github.com/spf13/viper"
)

func SetupConfig() {
	viper.SetEnvPrefix("dyntcp")
	viper.AutomaticEnv()

	viper.SetDefault("consul.addr", "http://127.0.0.1:8500")
	viper.SetDefault("gomaxprocs", runtime.NumCPU())
}
