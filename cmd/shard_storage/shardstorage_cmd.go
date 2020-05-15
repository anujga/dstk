package main

import (
	"fmt"
	"github.com/anujga/dstk/pkg/sharder"
	"github.com/spf13/viper"
)

func main() {
	setConfig()
	sharder.StartShardStorageServer(viper.GetInt32("sharder_storage.port"))
}

func setConfig() {
	viper.SetConfigName("SharderConfig")
	//viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
