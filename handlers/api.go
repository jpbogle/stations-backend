package handlers

import (
	"stations/utils"
	"net/http"
)

func init() {
	apiHandler := utils.CreateHandler()
	RootHandler.AddHandler("/api", apiHandler)

	apiHandler.Get("/demo", func(ctx *utils.Context) {
		ctx.Respond("<h1>Here's an api route demo!</h1>", http.StatusOK)
	})

	apiHandler.Get("/:stationName", func(ctx *utils.Context) {
		ctx.Respondf("<h1>Welcome to the station %s</h1>", http.StatusOK, ctx.Fields["stationName"])
	})

	registerUser(apiHandler)

	registerStation(apiHandler)
}
