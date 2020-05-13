package config

import (
	"log"

	"github.com/spf13/viper"
)

// RuntimeViper viper Instance
var RuntimeViper *viper.Viper

func init() {
	RuntimeViper = viper.New()
	RuntimeViper.SetEnvPrefix("service")                           // 将自动大写
	RuntimeViper.BindEnv("port")                                   // SERVICE_PORT
	RuntimeViper.BindEnv("mysql.service", "SERVICE_MYSQL_SERVICE") // SERVICE_MYSQL_SERVICE
	RuntimeViper.BindEnv("mysql.port", "SERVICE_MYSQL_PORT")       // SERVICE_MYSQL_PORT
	RuntimeViper.SetConfigType("yaml")
	RuntimeViper.SetConfigName("config")    // name of config file (without extension)
	RuntimeViper.AddConfigPath("./config/") // path to look for the config file in
	err := RuntimeViper.ReadInConfig()      // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s", err)
	}
}
