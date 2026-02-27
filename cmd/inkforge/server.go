package main

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/InkForge/internal/api"
	"github.com/playwright-community/playwright-go"
)

var (
	port string
	host string
)

func StartServer() {
	err := playwright.Install()
	if err != nil {
		panic(err)
	}

	eng := gin.Default()
	api.SetRouter(eng)

}
