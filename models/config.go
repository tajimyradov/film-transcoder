package models

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	FilmsDB  Postgres `mapstructure:"films_db"`
	Minio    Minio    `mapstructure:"minio"`
	RabbitMQ RabbitMQ `mapstructure:"rabbitmq"`
}

type Postgres struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
}

type Minio struct {
	EndPoint           string `mapstructure:"endpoint"`
	AccessKey          string `mapstructure:"access_key"`
	SecretAccessKey    string `mapstructure:"secret_key"`
	UseSsl             bool   `mapstructure:"use_ssl"`
	OrigianlFileBucket string `mapstructure:"original_file_bucket"`
	HLSFileBucket      string `mapstructure:"hls_file_bucket"`
}

type RabbitMQ struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Url       string `mapstructure:"url"`
	QueueName string `mapstructure:"queue_name"`
}

func NewAppConfig(configFile string) (*AppConfig, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadConfig(configFile string) (*AppConfig, error) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var appConfig AppConfig
	err = viper.Unmarshal(&appConfig)
	if err != nil {
		return nil, err
	}

	return &appConfig, err
}
