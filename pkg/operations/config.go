package operations

import (
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

func SetupConfig() {
	viper.SetEnvPrefix("dyntcp")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("consul.addr", "http://127.0.0.1:8500")
	viper.SetDefault("gomaxprocs", runtime.NumCPU())
	viper.SetDefault("ports.min", 1024)
	viper.SetDefault("ports.max", 49151)
	viper.SetDefault("peek_listener", "")

	viper.SetDefault("allowed_listeners", "local://127.0.0.1,any://0.0.0.0")
	allowedListeners := make(map[string]string)
	for _, l := range strings.Split(viper.GetString("allowed_listeners"), ",") {
		elements := strings.Split(l, "://")
		allowedListeners[elements[0]] = elements[1]
	}
	viper.Set("allowed_listeners", allowedListeners)
}
