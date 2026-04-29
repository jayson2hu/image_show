package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayson2hu/image-show/web"
)

func registerWebRoutes(r *gin.Engine) {
	dist, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(dist))

	r.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(dist, path); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		if _, err := fs.Stat(dist, "index.html"); err == nil {
			c.FileFromFS("index.html", http.FS(dist))
			return
		}

		c.Status(http.StatusNotFound)
	})
}
