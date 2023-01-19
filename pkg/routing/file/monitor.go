package file

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"dill/pkg/controller"
)

func MonitorServices(c chan<- *controller.RoutingTable) {
	v := viper.New()
	cfg := readRoutingConfig(v)
	rt := BuildRoutingTable(cfg)
	c <- &rt

	if viper.GetBool("routing.file.watch") {
		v.OnConfigChange(func(e fsnotify.Event) {
			rc := readRoutingConfig(viper.New())
			t := BuildRoutingTable(rc)
			c <- &t
		})
		v.WatchConfig()
	}
}
