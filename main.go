package main

import (
	"github.com/gin-gonic/gin"
	"github.com/yunarta/terraform-state-proxy/internal"
	"log"
	"net/http"
)

func main() {
	var config, err = internal.TestGetConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	r := gin.Default()

	bitbucket := internal.NewBitbucketHandler(config.Bitbucket.Server)

	r.GET("/bitbucket/:project/:repo/*path", bitbucket.Get)
	r.POST("/bitbucket/:project/:repo/*path", bitbucket.Post)

	gitea := internal.NewGiteaHandler(config.Gitea.Server)
	r.GET("/gitea/:project/:repo/*path", gitea.Get)
	r.POST("/gitea/:project/:repo/*path", gitea.Post)

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "Method not allowed")
	})

	r.Run(":8080")
}
