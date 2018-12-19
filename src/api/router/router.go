package router

import (
	"FunnyVoteGo/src/api/router/middleware"
	"FunnyVoteGo/src/api/v1"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	melody "gopkg.in/olahol/melody.v1"
)

// Load loads the middlewares, routes, handlers.
func Load(g *gin.Engine, mg *melody.Melody, mw ...gin.HandlerFunc) *gin.Engine {
	// Middlewares.
	g.Use(gin.Recovery())
	g.Use(gin.Logger())
	g.Use(middleware.NoCache)
	g.Use(middleware.Options)
	g.Use(middleware.Secure)
	g.Use(mw...)

	if viper.GetString("runmode") == "debug" {
		g.GET("/", func(c *gin.Context) {
			c.Redirect(301, "/swagger/index.html")
		})
		g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	//main router
	apiv1 := g.Group("/api/v1")
	apiv1.GET("/info", v1.GetContractInfo)
	apiv1.POST("/startvote", v1.StartVote)

	return g
}
