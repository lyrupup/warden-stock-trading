// Package config 使用 viper 读取 config.yaml 并支持环境变量覆盖（不硬编码）。
package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置聚合。
type Config struct {
	App       AppConfig      `mapstructure:"app"`
	JWT       JWTConfig      `mapstructure:"jwt"`
	Postgres  PostgresConfig `mapstructure:"postgres"`
	Redis     RedisConfig    `mapstructure:"redis"`
	Market    MarketConfig   `mapstructure:"market"`
	RateLimit RateLimit      `mapstructure:"ratelimit"`
	Timeout   TimeoutConfig  `mapstructure:"timeout"`
	Screen    ScreenConfig   `mapstructure:"screen"`
	// ConfigEncKey 敏感配置 AES-GCM 主密钥，来自环境变量 CONFIG_ENC_KEY，不入库不入仓。
	ConfigEncKey string `mapstructure:"config_enc_key"`
}

type AppConfig struct {
	Port           int    `mapstructure:"port"`
	SingleUserMode bool   `mapstructure:"single_user_mode"`
	Env            string `mapstructure:"env"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type PostgresConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"db"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MarketConfig struct {
	Provider     string `mapstructure:"provider"`
	GotdxMaxConn int    `mapstructure:"gotdx_max_conn"`
}

type RateLimit struct {
	RPS   int `mapstructure:"rps"`
	Burst int `mapstructure:"burst"`
}

type TimeoutConfig struct {
	Seconds int `mapstructure:"seconds"`
}

// ScreenConfig 量化粗筛批处理参数。
// Concurrency：单次粗筛内并发拉取 K 线 / 评估规则的协程数，建议不超过 market.gotdx_max_conn。
// TimeoutSeconds：粗筛接口级超时（覆盖全局 timeout.seconds），可单独放宽以容纳批量行情拉取。
type ScreenConfig struct {
	Concurrency    int `mapstructure:"concurrency"`
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
}

// Load 读取配置。path 为空时按 ./config/config.yaml 与 ./config.yaml 搜索。
// 环境变量优先级最高（见 BACKEND.md §7.2）。
func Load(path string) (*Config, error) {
	v := viper.New()
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	setDefaults(v)
	bindEnv(v)

	if err := v.ReadInConfig(); err != nil {
		// 配置文件缺失不算致命（可纯靠环境变量），仅当解析错误才返回。
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.single_user_mode", true)
	v.SetDefault("app.env", "dev")
	v.SetDefault("jwt.secret", "change_me")
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "postgres")
	v.SetDefault("postgres.password", "postgres")
	v.SetDefault("postgres.db", "warden")
	v.SetDefault("postgres.sslmode", "disable")
	v.SetDefault("postgres.max_open_conns", 20)
	v.SetDefault("postgres.max_idle_conns", 10)
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("market.provider", "gotdx")
	v.SetDefault("market.gotdx_max_conn", 8)
	v.SetDefault("ratelimit.rps", 100)
	v.SetDefault("ratelimit.burst", 200)
	v.SetDefault("timeout.seconds", 10)
	v.SetDefault("screen.concurrency", 8)
	v.SetDefault("screen.timeout_seconds", 60)
	v.SetDefault("config_enc_key", "")
}

// bindEnv 显式绑定 BACKEND.md §7.2 约定的扁平环境变量到嵌套配置键。
func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("app.port", "APP_PORT")
	_ = v.BindEnv("app.single_user_mode", "SINGLE_USER_MODE")
	_ = v.BindEnv("app.env", "APP_ENV")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("postgres.host", "PG_HOST")
	_ = v.BindEnv("postgres.port", "PG_PORT")
	_ = v.BindEnv("postgres.user", "PG_USER")
	_ = v.BindEnv("postgres.password", "PG_PASSWORD")
	_ = v.BindEnv("postgres.db", "PG_DB")
	_ = v.BindEnv("postgres.sslmode", "PG_SSLMODE")
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("market.provider", "MARKET_PROVIDER")
	_ = v.BindEnv("market.gotdx_max_conn", "MARKET_GOTDX_MAX_CONN")
	_ = v.BindEnv("screen.concurrency", "SCREEN_CONCURRENCY")
	_ = v.BindEnv("screen.timeout_seconds", "SCREEN_TIMEOUT_SECONDS")
	_ = v.BindEnv("config_enc_key", "CONFIG_ENC_KEY")
}
