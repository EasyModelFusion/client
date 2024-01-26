package utils

import (
	"github.com/easy-model-fusion/client/internal/app"
	"github.com/spf13/viper"
)

// GetViperItem Store the key data into the target
func GetViperItem(key string, target interface{}) error {
	logger := app.L().WithTime(false)
	if err := viper.UnmarshalKey(key, target); err != nil {
		logger.Error("Error reading config file:" + err.Error())
		return err
	}
	return nil
}

// WriteViperConfig Attempt to write the configuration file
func WriteViperConfig() error {
	logger := app.L().WithTime(false)
	if err := viper.WriteConfig(); err != nil {
		logger.Error("Error writing to config file:" + err.Error())
		return err
	}
	return nil
}
