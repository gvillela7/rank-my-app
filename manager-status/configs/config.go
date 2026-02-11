package config

import (
	"errors"

	"github.com/spf13/viper"
)

var cfg *config

type config struct {
	API      APIConfig
	DBMongo  DBMongo
	RabbitMQ RabbitMQConfig
}

type APIConfig struct {
	Port          string
	Environment   string
	Host          string
	Origin        string
	Documentation string
	LogDir        string
	TimeZone      string
}

type DBMongo struct {
	URI      string
	Database string
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	VHost    string
}

func init() {
	//Service
	viper.SetDefault("api.port", "8000")
	viper.SetDefault("api.environment", "dev")
	viper.SetDefault("api.host", "localhost")
	viper.SetDefault("api.origin", "*")
	viper.SetDefault("api.documentation", "http://localhost:8001/swagger/index.html")
	viper.SetDefault("api.timezone", "America/Sao_Paulo")

	//MongoDB
	viper.SetDefault("mongo.uri", "")
	viper.SetDefault("mongo.database", "")

	//RabbitMQ
	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", "5672")
	viper.SetDefault("rabbitmq.username", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")

}

func Load(viperPath ...string) error {
	if len(viperPath) > 0 {
		viper.AddConfigPath(viperPath[0])
	} else {
		viper.AddConfigPath(".")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	cfg = new(config)

	cfg.API = APIConfig{
		Port:          viper.GetString("api.port"),
		Environment:   viper.GetString("api.environment"),
		Host:          viper.GetString("api.host"),
		Origin:        viper.GetString("api.origins"),
		TimeZone:      viper.GetString("api.timezone"),
		Documentation: viper.GetString("api.documentation"),
	}

	cfg.DBMongo = DBMongo{
		URI:      viper.GetString("mongo.uri"),
		Database: viper.GetString("mongo.database"),
	}

	cfg.RabbitMQ = RabbitMQConfig{
		Host:     viper.GetString("rabbitmq.host"),
		Port:     viper.GetString("rabbitmq.port"),
		Username: viper.GetString("rabbitmq.username"),
		Password: viper.GetString("rabbitmq.password"),
		VHost:    viper.GetString("rabbitmq.vhost"),
	}

	return nil
}

func GetAPIConfig() APIConfig {
	return cfg.API
}

func GetDBMongo() DBMongo {
	return cfg.DBMongo
}

func GetRabbitMQConfig() RabbitMQConfig {
	return cfg.RabbitMQ
}
