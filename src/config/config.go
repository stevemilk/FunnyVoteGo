package config

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/glog"
	"github.com/spf13/viper"
)

// Config define
type Config struct {
	Name string
}

// Init prepare config
func Init(cfg string) error {
	c := Config{
		Name: cfg,
	}

	if err := c.initConfig(); err != nil {
		return err
	}

	c.watchConfig()

	return nil
}

func (c *Config) initConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name)
	} else {
		viper.AddConfigPath("conf")
		viper.SetConfigName("config")
	}
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()            // Read Env value
	viper.SetEnvPrefix("APISERVER") // Env prefix : APISERVER
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil { // viper read config
		return err
	}

	return nil
}

// watch config change
func (c *Config) watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		glog.Infof("Config file changed: %s", e.Name)
	})
}
