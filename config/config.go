package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	TgBot TgBot    `mapstructure:",squash"`
	PSQL  Postgres `mapstructure:",squash"`
}

type Postgres struct {
	PgUser     string
	PgPassword string
	PgHost     string
	PgPort     string
	PgDb       string
}

type TgBot struct {
	Token string `mapstructure:"bot_token"`
}

func (c *Config) InitCfg() error {
	viper.AddConfigPath("./")
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = viper.Unmarshal(&c)

	return err
}
