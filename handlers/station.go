package handlers

import (
	"encoding/json"
	"net/http"
	"stations/controllers"
	"stations/entities"
	"stations/utils"
	// "github.com/dgrijalva/jwt-go"
	"fmt"
	// "os"
	// "io"
	// "io/ioutil"
	// "time"
	// "crypto/ecdsa"
	// "flag"
	// "encoding/pem"
)


func registerStation(apiHandler *utils.Handler) {

	// Tune in with a username
	apiHandler.Get("/:username/:stationName/ws/:listener", func(ctx *utils.Context) {
		station, err := controllers.GetStation(ctx.Fields["username"], ctx.Fields["stationName"])
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}
		isAdmin := station.Creator == ctx.Fields["listener"]
		for _, shallowUser := range station.Admins {
			if shallowUser.Username == ctx.Fields["listener"] {
				isAdmin = true
			}
		}
		if isAdmin {
			ctx.Broadcast(station, "New Admin", "Someone is now controlling the station!")
		} else {
			ctx.Broadcast(station, "New Listener", "Someone tuned in!")
		}
		ctx.OpenWebsocket(station, isAdmin)
	})

	// Tune in without a username
	apiHandler.Get("/:username/:stationName/ws/", func(ctx *utils.Context) {
		station, err := controllers.GetStation(ctx.Fields["username"], ctx.Fields["stationName"])
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}
		ctx.OpenWebsocket(station, false)
	})

	// Get a station's info
	apiHandler.Get("/:username/:stationName", func(ctx *utils.Context) {
		station, err := controllers.GetStation(ctx.Fields["username"], ctx.Fields["stationName"])
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

////////////////////////////// JWT stuff ///////////////////////////

		// token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		//     "foo": "bar",
		//     "nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		// })

		// var privateKey *ecdsa.PrivateKey
		// const privKeyPath string = "./stations.p8"
		// var privateKeyBytes []byte

		// privateKeyBytes, err = ioutil.ReadFile(privKeyPath)
		 //    if err != nil {
		 //        log.Print(err)
		 //    }
		// privateKey, err = jwt.ParseECPrivateKeyFromPEM(privateKeyBytes)

		// // Sign and get the complete encoded token as a string using the secret
		// tokenString, err := token.SignedString(privateKey)

		// fmt.Println(tokenString, err)

////////////////////////////////////////////////////////////////////

		station.AppleMusicToken = "fkorkf"

		resp := entities.CreateStationResponse{
			Station: station,
		}
		ctx.Broadcast(station, "New Listener", "Someone got station info")
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Create a new station
	apiHandler.Post("/stations/create", func(ctx *utils.Context) {
		var createStationRequest entities.CreateStationRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&createStationRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}

		station, err := controllers.CreateStation(&createStationRequest)
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

		resp := entities.CreateStationResponse{
			Station: station,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Get all songs in a station
	apiHandler.Get("/:username/:stationName/shuffle", func(ctx *utils.Context) {
		station, err := controllers.ShuffleDefaults(ctx.Fields["username"], ctx.Fields["stationName"]);
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}
		message := fmt.Sprintf("Station shuffled...")
		resp := entities.CreateStationResponse{
			Station: station,
		}

		ctx.Broadcast(station, "Shuffle", message)
		ctx.RespondJson(resp, http.StatusOK)
	})


	// Add a song to a station
	apiHandler.Post("/:username/:stationName/songs/add", func(ctx *utils.Context) {
		var createSongRequest entities.CreateSongRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&createSongRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}

		addSongRequest := entities.AddSongRequest{
			Creator:     ctx.Fields["username"],
			StationName: ctx.Fields["stationName"],
			Song:        createSongRequest,
		}

		station, err := controllers.AddSong(&addSongRequest)
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}
		message := fmt.Sprintf("Added %s - %s", createSongRequest.Title, createSongRequest.Artist)
		resp := entities.CreateStationResponse{
			Station: station,
		}

		ctx.Broadcast(station, "Song Added", message)
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Closes a station
	apiHandler.Get("/:username/:stationName/reset", func(ctx *utils.Context) {
		ctx.CloseWebsocket()
	})

	// Add an administrator to a station
	apiHandler.Post("/:username/:stationName/admin/add", func(ctx *utils.Context) {
		var addAdminRequest entities.AddAdminRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&addAdminRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}
		addAdminRequest.Creator = ctx.Fields["username"]
		addAdminRequest.StationName = ctx.Fields["stationName"]

		station, err := controllers.AddAdmin(&addAdminRequest)
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

		resp := entities.CreateStationResponse{
			Station: station,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Play next song
	apiHandler.Post("/:username/:stationName/play/next", func(ctx *utils.Context) {
		station, err := controllers.PlayNext(ctx.Fields["username"], ctx.Fields["stationName"])
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}
		message := fmt.Sprintf("Now Playing %s - %s", station.Playing.Song.Title, station.Playing.Song.Artist)
		resp := entities.CreateStationResponse{
			Station: station,
		}
		ctx.Broadcast(station, "Next Song", message)
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Vote a song to a station
	apiHandler.Post("/:username/:stationName/songs/vote", func(ctx *utils.Context) {
		var voteRequest entities.VoteRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&voteRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}

		stationId, err := controllers.GetStationId(ctx.Fields["username"], ctx.Fields["stationName"])


		Song, err := controllers.GetSong(voteRequest.Source, voteRequest.SourceId)

		var isUpVote bool

		if voteRequest.Action == "upvote" {
			isUpVote = true
		} else {
			isUpVote = false
		}

		station, err := controllers.ChangeVote(stationId, Song.Id, isUpVote)
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}
		message := fmt.Sprintf("Vote changed on %s - %s", Song.Title, Song.Artist)
		resp := entities.CreateStationResponse{
			Station: station,
		}

		ctx.Broadcast(station, "Vote", message)
		ctx.RespondJson(resp, http.StatusOK)
	})

	apiHandler.Get("/token", func(ctx *utils.Context) {
		resp := entities.APIToken{
			Token: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlNNSlNCOUFHVVEifQ.eyJpc3MiOiIyRVhWREo4OE4yIiwiaWF0IjoxNTMzNTQwNjU0LCJleHAiOjE1MzM1NDE2NTR9.uDcT29g9hg6KVuCumvB-7z5mUuGFjZTsIPMtcY3BrFdCDQ6rb-oS9p5rQcSZQjmMv7XfLjSYzZ28KTvvlwnxLg",
		}
		ctx.RespondJson(resp, http.StatusOK)
	});
}
