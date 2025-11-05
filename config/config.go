package config

import "github.com/spf13/viper"

type Config struct {
	Server *ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func LoadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	if err := viper.ReadInConfig(); err != nil {
		panic("viper read config failed:" + err.Error())
	}

	cf := new(Config)

	if err := viper.Unmarshal(cf); err != nil {
		panic("viper unmarshal config failed:" + err.Error())
	}
	return cf
}
