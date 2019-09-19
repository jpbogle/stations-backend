package handlers

import (
	"encoding/json"
	"net/http"
	"stations/controllers"
	"stations/entities"
	"stations/utils"
	"fmt"
	"net/url"
	// "io/ioutil"
	"strings"
	b64 "encoding/base64"

)

func registerUser(apiHandler *utils.Handler) {
	// Fetch all users
	apiHandler.Get("/users", func(ctx *utils.Context) {
		users, err := controllers.GetUsers()
		if notFoundErr, ok := err.(*entities.NotFoundError); ok {
			ctx.Error(notFoundErr, http.StatusOK)
			return
		}
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

		resp := entities.GetUsersResponse{
			Users: users,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Fetch a user
	apiHandler.Get("/users/:username", func(ctx *utils.Context) {
		user, err := controllers.GetUser(ctx.Fields["username"])
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

		resp := entities.GetUserResponse{
			User: user,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Update any value of a user object
	apiHandler.Post("/users/update", func(ctx *utils.Context) {
		var updateUserRequest entities.UpdateUserRequest
		if err := json.NewDecoder(ctx.Req.Body).Decode(&updateUserRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}

		user, err := controllers.UpdateUser(&updateUserRequest)
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

		resp := entities.GetUserResponse{
			User: user,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Create a new user
	apiHandler.Post("/users/create", func(ctx *utils.Context) {
		var userRequest entities.CreateUserRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&userRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}

		user, err := controllers.CreateUser(&userRequest)
		if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}

		ctx.Session().Set("username", user.Username)

		resp := entities.GetUserResponse{
			User: user,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Login with a username and password
	apiHandler.Post("/users/login", func(ctx *utils.Context) {
		var loginUserRequest entities.LoginUserRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&loginUserRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}

		user, err := controllers.LoginUser(&loginUserRequest)
		fmt.Println(err)

		if notFoundErr, ok := err.(*entities.NotFoundError); ok {
			ctx.Error(notFoundErr, http.StatusUnauthorized)
			return
		} else if err != nil {
			ctx.Error(err, http.StatusInternalServerError)
			return
		}


		if user == nil { // bad credentials supplied
			ctx.Error(&entities.Error{Msg: "Bad credentials"}, http.StatusUnauthorized)
			return
		}

		ctx.Session().Set("username", user.Username)

		resp := entities.GetUserResponse{
			User: user,
		}
		ctx.RespondJson(resp, http.StatusOK)
	})

	// Add an account to a user
	apiHandler.Post("/users/:username/accounts", func(ctx *utils.Context) {
		var addAccountRequest entities.AddAccountRequest

		if err := json.NewDecoder(ctx.Req.Body).Decode(&addAccountRequest); err != nil {
			ctx.Error(err, http.StatusBadRequest)
			return
		}
		username := ctx.Fields["username"]
		clientId := "bec7cb795af042689409eb98a961e77e"
        clientSecret := "de284712da9d4cff85f5e537d055c9b5"

		// bytesRepresentation, err := json.Marshal(message)
		if addAccountRequest.Source == "spotify" {
			client := &http.Client{}
			form := url.Values{}
			form.Set("code", addAccountRequest.Code)
			form.Add("grant_type", "authorization_code")
			form.Add("redirect_uri", "http://localhost:4000/dashboard")

			req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", "Basic " + b64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret)))

			res, err := client.Do(req)
			defer res.Body.Close()
			fmt.Println(res.Body)
			var tokens entities.Tokens
			if err := json.NewDecoder(res.Body).Decode(&tokens); err != nil {
				ctx.Error(err, http.StatusBadRequest)
				return
			}

			var user *entities.User
			user, err = controllers.AddAccount(username, "spotify", &tokens)
			if err != nil {
				ctx.Error(err, http.StatusBadRequest)
				return
			}

			resp := entities.GetUserResponse{
				User: user,
			}
			ctx.RespondJson(resp, http.StatusOK)
			return
		}
		ctx.Respond("<h1>Here's an api route demo!</h1>", http.StatusOK)
	})

	// Add an account to a user
	apiHandler.Get("/users/:username/refresh/:source", func(ctx *utils.Context) {

		username := ctx.Fields["username"]

		clientId := "bec7cb795af042689409eb98a961e77e"
        clientSecret := "de284712da9d4cff85f5e537d055c9b5"

		// bytesRepresentation, err := json.Marshal(message)
		if  ctx.Fields["source"] == "spotify" {
			oldTokens, err := controllers.GetUserTokens(username)
			if err !=nil {
				ctx.Error(err, http.StatusBadRequest)
				return
			}
			var refreshToken string
			for _, v := range oldTokens {
			    if v.Source == "spotify" {
			    	refreshToken = v.RefreshToken
			    }
			}

			client := &http.Client{}
			form := url.Values{}
			form.Set("refresh_token", refreshToken)
			form.Add("grant_type", "refresh_token")

			req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", "Basic " + b64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret)))

			res, err := client.Do(req)
			defer res.Body.Close()
			fmt.Println(res.Body)
			var tokens entities.Tokens
			if err := json.NewDecoder(res.Body).Decode(&tokens); err != nil {
				ctx.Error(err, http.StatusBadRequest)
				return
			}

			var user *entities.User
			user, err = controllers.RefreshToken(username, "spotify", tokens.AccessToken)
			if err != nil {
				ctx.Error(err, http.StatusBadRequest)
				return
			}

			resp := entities.GetUserResponse{
				User: user,
			}
			ctx.RespondJson(resp, http.StatusOK)
			return
		}
		ctx.Respond("<h1>Here's an api route demo!</h1>", http.StatusOK)
	})

	// Clear the client's session
	apiHandler.Post("/users/logout", func(ctx *utils.Context) {
		ctx.DestroySession()
		ctx.Respond("", http.StatusOK)
	})

	// Get the client's session
	apiHandler.Get("/users/session", func(ctx *utils.Context) {
		resp := entities.GetUserResponse{
			User: ctx.GetAuthenticatedUser(),
		}
		ctx.RespondJson(resp, http.StatusOK)
	})
}
