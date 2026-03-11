package main

import (
	"flag"
	"fmt"
	"mytemplate/internal/config"
	"mytemplate/internal/database"
	"mytemplate/internal/global"
	"mytemplate/internal/handler"
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	c, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("config load error %v", err))
	}
	global.AppConfig = c

	database.Setup(*c) // mysql
	//cache.Setup(*c)    // redis

	go startMetric()
	fmt.Printf("parser config Host=%s, Port=%d\n", c.Host, c.Port)

	handler.Setup()
}

func loadConfig() (*config.Config, error) {
	var configFile string
	flag.StringVar(&configFile, "f", "etc/activity-manager-api.yaml", "public config file path")
	flag.Parse()

	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	fmt.Printf("now env: %s\n", env)

	v := viper.New()
	v.SetConfigType("yaml")

	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("load public config failed: %v", err)
	}

	envConfigFile := fmt.Sprintf("etc/activity-manager-api.%s.yaml", env)
	if _, err := os.Stat(envConfigFile); err == nil {
		fmt.Printf("load env config: %s\n", envConfigFile)
		v.SetConfigFile(envConfigFile)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge env config failed: %v", err)
		}
	} else {
		fmt.Printf("env config file not found, using public config only: %s\n", envConfigFile)
	}

	var c config.Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %v", err)
	}
	return &c, nil
}

func startMetric() {
	// start metrics server
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	serveMux.Handle("/metrics", promhttp.Handler())

	// start metrics server
	http.ListenAndServe(":8081", serveMux)
}
