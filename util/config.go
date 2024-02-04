package util

import "github.com/spf13/viper"

// Config stores all config settings of the app
// Values are read using viper
type Config struct {
	DBDriver     string `mapstructure:"DB_DRIVER"`
	DBSource     string `mapstructure:"DB_SOURCE"`
	SeverAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig reads config settings from file/ env variables
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
