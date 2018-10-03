package handlers

import (
	"encoding/json"
	"net/http"
	"stations/controllers"
	"stations/entities"
	"stations/utils"
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
