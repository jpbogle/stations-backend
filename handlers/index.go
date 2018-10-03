package handlers

import (
	"net/http"
	"stations/utils"
)


var RootHandler = utils.CreateRootHandler()

func init() {
	RootHandler.Get("", func(ctx *utils.Context) {
		ctx.Respond("<h1>Welcome to index</h1>", http.StatusOK)
	})

	RootHandler.Get("/health", func(ctx *utils.Context) {
		ctx.Respond("Hello Stations!", http.StatusOK)
	})
}
