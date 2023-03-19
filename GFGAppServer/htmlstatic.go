package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func setup(router *gin.Engine) {
	router.Static("/assets", "./assets")
	router.Static("/images", "./images")
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"content": "Login page",
		})
	})

	router.GET("/account.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "account.html", gin.H{
			"content": "User account page",
		})
	})
}
