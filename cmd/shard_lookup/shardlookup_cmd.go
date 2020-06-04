package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	setConfig()
	jobsStr := viper.GetStringSlice("sharder_lookup.jobs")
	jobs := make([]int64, 0)
	for job := range jobsStr {
		jobs = append(jobs, int64(job))
	}
	//sharder.StartShardLookupService(viper.GetInt32("sharder_lookup.port"), jobs)
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
