package config

import "github.com/spf13/viper"

type Config struct {
	Server *ServerConfig `mapstructure:"server"`
	log    *LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`     // 日志等级
	Format   string `mapstructure:"format"`    // 日志格式
	Path     string `mapstructure:"path"`      // 日志文件路径
	FileName string `mapstructure:"file_name"` // 日志文件名
	MaxSize  int    `mapstructure:"max_size"`  // 单个日志文件最大尺寸，单位MB
	MaxAge   int    `mapstructure:"max_age"`   // 日志文件最大保存天数
	Compress bool   `mapstructure:"compress"`  // 是否压缩日志文件
	Stdout   bool   `mapstructure:"stdout"`    // 是否输出到控制台
}

type ScheduleConfig struct {
	WithSeconds bool `mapstructure:"with_seconds"` // 是否启用秒级调度
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
