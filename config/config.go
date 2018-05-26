package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

var Config taskMail

type taskMail struct {
	SMTPHost string   `mapstructure:"host"`
	SMTPPort int      `mapstructure:"port"`
	SMTPPass string   `mapstructure:"password"`
	To       []string `mapstructure:"to"`
	From     string   `mapstructure:"from"`
	FromName string   `mapstructure:"from_name"`
}

func (config taskMail) Validate() error {
	if len(config.SMTPPass) == 0 {
		return errors.New("You must specify the password")
	}
	return nil
}

func LoadConfig(configPaths ...string) error {
	v := viper.New()
	v.SetConfigName("taskmail")
	v.SetConfigType("yaml")
	v.SetEnvPrefix("taskmail")
	v.AutomaticEnv()
	v.SetDefault("port", 587)

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read the configuration file: %s", err)
	}

	if err := v.Unmarshal(&Config); err != nil {
		return err
	}

	return Config.Validate()
}
