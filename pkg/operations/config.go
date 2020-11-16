package operations

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Configuration file path")
}

func SetupConfig() {
	viper.SetEnvPrefix("dyntcp")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("consul.address", "http://127.0.0.1:8500")
	viper.SetDefault("runtime.gomaxprocs", runtime.NumCPU())
	viper.SetDefault("listeners.ports_min", 1024)
	viper.SetDefault("listeners.ports_max", 49151)
	viper.SetDefault("peek.listener", "")
	viper.SetDefault(
		"listeners.allowed",
		map[string]string{"local": "127.0.0.1", "any": "0.0.0.0"},
	)

	flag.Parse()
	if configPath != "" {
		viper.SetConfigFile(configPath)
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			fmt.Printf("config error: %s\n", err)
			os.Exit(1)
		}
	}
}
