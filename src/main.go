package main

import (
	"FunnyVoteGo/src/api/router"
	"FunnyVoteGo/src/config"
	"FunnyVoteGo/src/model"
	"flag"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/glog"
	"github.com/spf13/viper"
)

var (
	cfg = flag.String("config", "", "FunnyVoteGo config file path.")
)

func main() {

	if cpu := runtime.NumCPU(); cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}
	// init config
	if err := config.Init(*cfg); err != nil {
		panic(err)
	}
	// set gin mode
	gin.SetMode(viper.GetString("runmode"))

	// Create the Gin engine.
	g := gin.New()

	middlewares := []gin.HandlerFunc{}

	model.InitDataBase()

	// Routes.
	router.Load(
		// Cores.
		g,
		nil,

		// Middlwares.
		middlewares...,
	)

	glog.Infof("Start to listening the incoming requests on http address: %s", viper.GetString("addr"))

	glog.Infof(http.ListenAndServe(viper.GetString("addr"), g).Error())

}
