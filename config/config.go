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
	JWT      JWTConfig      `mapstructure:"jwt"`
	Profile  ProfileConfig  `mapstructure:"profile"`
	Console  ConsoleConfig  `mapstructure:"console"`
	Seed     SeedConfig     `mapstructure:"seed"`
}

// ConsoleConfig svr-console 下游服务配置（订阅管控）
type ConsoleConfig struct {
	BaseURL     string `mapstructure:"base_url"`
	InternalKey string `mapstructure:"internal_key"`
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// ProfileConfig svr-profile 下游服务配置（用户管控编排）
type ProfileConfig struct {
	BaseURL     string `mapstructure:"base_url"`
	InternalKey string `mapstructure:"internal_key"`
}

// SeedConfig 初始超级管理员种子配置（首次启动自动创建）
type SeedConfig struct {
	SuperAdminUsername string `mapstructure:"super_admin_username"`
	SuperAdminPassword string `mapstructure:"super_admin_password"`
	SuperAdminEmail    string `mapstructure:"super_admin_email"`
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

	// jwt
	bindEnv("jwt.secret", "JWT_SECRET")

	// profile（下游 svr-profile）
	bindEnv("profile.base_url", "PROFILE_BASE_URL")
	bindEnv("profile.internal_key", "PROFILE_INTERNAL_KEY")

	// console（下游 svr-console）
	bindEnv("console.base_url", "CONSOLE_BASE_URL")
	bindEnv("console.internal_key", "CONSOLE_INTERNAL_KEY")

	// seed（初始超管，仅首次启动生效）
	bindEnv("seed.super_admin_username", "SEED_SUPER_ADMIN_USERNAME")
	bindEnv("seed.super_admin_password", "SEED_SUPER_ADMIN_PASSWORD")
	bindEnv("seed.super_admin_email", "SEED_SUPER_ADMIN_EMAIL")
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

	// jwt
	viper.SetDefault("jwt.secret", "svr-admin-jwt-secret-change-me")

	// profile
	viper.SetDefault("profile.base_url", "")
	viper.SetDefault("profile.internal_key", "")

	// console
	viper.SetDefault("console.base_url", "")
	viper.SetDefault("console.internal_key", "")

	// seed
	viper.SetDefault("seed.super_admin_username", "admin")
	viper.SetDefault("seed.super_admin_password", "")
	viper.SetDefault("seed.super_admin_email", "admin@nostack.local")
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
