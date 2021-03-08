package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Options defines the configuration options.
type Options struct {
	EnvVarEnabled bool
	EnvVarPrefix  string

	ConfigPath     string
	ConfigFileName string
}

// Init initializes service configuration
func Init(opts *Options) error {
	// For environment variables
	if opts.EnvVarEnabled {
		viper.SetEnvPrefix(opts.EnvVarPrefix)
		viper.AutomaticEnv()
		replacer := strings.NewReplacer(".", "_")
		viper.SetEnvKeyReplacer(replacer)
	}

	viper.SetConfigName(opts.ConfigFileName)

	configPath := opts.ConfigPath
	if configPath == "" {
		envConfig := os.Getenv(opts.EnvVarPrefix + "_CFG_PATH")
		if envConfig == "" {
			configPath = "config" + string(filepath.Separator)
		} else {
			configPath = envConfig
		}
	}

	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	actives := viper.GetStringSlice("service.profiles.active")
	for _, active := range actives {
		activeConfigName := opts.ConfigFileName + "_" + active
		viper.SetConfigName(activeConfigName)
		err := viper.MergeInConfig()
		if err != nil {
			return err
		}
	}

	return nil
}
