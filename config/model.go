package config

type Config struct {
	Server   *ServerConfig   `mapstructure:"server"`
	Log      *LogConfig      `mapstructure:"log"`
	Schedule *ScheduleConfig `mapstructure:"schedule"`
	DB       *DBConfig       `mapstructure:"db"`
	JWT      *JWTConfig      `mapstructure:"jwt"`
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
	WithSeconds    bool   `mapstructure:"with_seconds"`    // 是否启用秒级调度
	WorkDir        string `mapstructure:"work_dir"`        // 任务工作目录
	GoroutinesSize int    `mapstructure:"goroutines_size"` // 任务执行协程池大小
}

type DBConfig struct {
	Addr string `mapstructure:"addr"` // 数据库地址
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"` // JWT密钥
}

func GetScheduleConfig(cf *Config) *ScheduleConfig {
	return cf.Schedule
}

func GetServerConfig(cf *Config) *ServerConfig {
	return cf.Server
}

func GetLogConfig(cf *Config) *LogConfig {
	return cf.Log
}

func GetDBConfig(cf *Config) *DBConfig {
	return cf.DB
}

func GetJWTConfig(cf *Config) *JWTConfig {
	return cf.JWT
}
