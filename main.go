package main

import (
	"gan/gan"
	"net/http"
)

func main() {
	r := gan.Default()
	// test GET
	r.GET("/", func(c *gan.Context) {
		c.String(http.StatusOK, "Hello from get, %v\n", c.Query("name"))
	})

	// test POST
	r.POST("/post", func(c *gan.Context) {
		c.String(http.StatusOK, "Hello from post, %v", c.PostForm("name"))
	})

	// test router params
	r.GET("/router/:name/*details", func(c *gan.Context) { //support named params & wildcards
		c.String(http.StatusOK, "Hello from router params, %v, %v", c.Param("name"), c.Param("details"))
	})

	// test Group
	g := r.Group("/people")
	g.GET("/:name", func(c *gan.Context) {
		c.String(http.StatusOK, "Hello from group, %v", c.Param("name"))
	})

	// test middleware
	auth := func(c *gan.Context) {
		username := c.Query("username")
		if username != "admin" {
			c.AbortWithError(http.StatusUnauthorized, "unauthorized!")
		}
	}
	g = r.Group("/auth")
	g.Use(auth)
	g.GET("/login", func(c *gan.Context) {
		c.String(http.StatusOK, "Hello from middleware, admin!")
	})

	// test panic recovery
	r.GET("/panic", func(c *gan.Context) {
		names := "stan"
		c.String(http.StatusOK, names[:5])
	})

	_ = r.Run(":8888")
}
