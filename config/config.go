package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release / test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

// DSN 返回数据库连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName,
	)
}

// 命令行参数
var (
	cfgFile string
)

func init() {
	pflag.StringVarP(&cfgFile, "config", "c", "", "配置文件路径 (默认查找 ./config.yaml)")
}

// bindEnv 绑定环境变量
func bindEnv(key, envKey string) {
	_ = viper.BindEnv(key, envKey)
}

// setupBindings 批量绑定环境变量
func setupBindings() {
	// server
	bindEnv("server.port", "SERVER_PORT")
	bindEnv("server.mode", "SERVER_MODE")

	// database (环境变量前缀 DB_)
	bindEnv("database.host", "DB_HOST")
	bindEnv("database.port", "DB_PORT")
	bindEnv("database.user", "DB_USER")
	bindEnv("database.password", "DB_PASSWORD")
	bindEnv("database.dbname", "DB_NAME")
}

// setDefaults 设置默认值
func setDefaults() {
	// server
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	// database
	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "svr_admin")
}

// Load 加载配置
// 优先级: 命令行参数 > 环境变量 > 配置文件 > 默认值
// 必须在 ParseFlags() 之后调用
func Load() *Config {
	// 1. 设置默认值
	setDefaults()

	// 2. 绑定环境变量
	setupBindings()

	// 3. 读取配置文件
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	// 4. 自动读取环境变量
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 5. 读取配置文件（文件不存在不报错，兼容纯环境变量模式）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Sprintf("解析配置文件失败: %v", err))
		}
	}

	// 6. 支持环境特定配置覆盖: config.local.yaml
	if cfgFile == "" {
		viper.SetConfigName("config.local")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		_ = viper.MergeInConfig()
	}

	// 7. 将配置反序列化到结构体
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("配置反序列化失败: %v", err))
	}

	return &cfg
}

// ParseFlags 解析命令行参数
// 必须在 Load() 之前调用
func ParseFlags() {
	pflag.Parse()
}

// ConfigFile 返回当前使用的配置文件路径（用于日志输出）
func ConfigFile() string {
	return viper.ConfigFileUsed()
}
