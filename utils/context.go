package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"stations/controllers"
	"stations/entities"
	"net"
	"bufio"
)

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	Hijack() (net.Conn, *bufio.ReadWriter, error)
}

//Context implements ResponseWriter which is http.ResponseWriter with Hijacking
type Context struct {
	Res    ResponseWriter
	Req    *http.Request
	Fields map[string]string
}

//////////////////////////////////////////////////////////////////////////////
//
// Context methods
//
//////////////////////////////////////////////////////////////////////////////

// Converts a JSON byte array to a well-formatted JSON byte array
func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// Gets the client's session if it exists, or creates a new session
func (ctx *Context) Session() Session {
	return globalSessions.GetSession(ctx)
}

// Destroys the client's session
func (ctx *Context) DestroySession() {
	globalSessions.DestroySession(ctx)
}

// Fetches the user entity for the currently authenticated user
func (ctx *Context) GetAuthenticatedUser() *entities.User {
	username := ctx.Session().Get("username")
	if username == nil {
		return nil
	}
	user, _ := controllers.GetUser(username.(string))
	return user
}

// Writes content to the context's http response and responds with http status
func (ctx *Context) Respond(content string, status int) {
	ctx.Res.WriteHeader(status)
	fmt.Fprintf(
		ctx.Res,
		"%s",
		content,
	)
}

// Formats according to the format specified and writes to the context's http response
// and responds with http.StatusOK
func (ctx *Context) Respondf(format string, status int, a ...interface{}) {
	ctx.Res.WriteHeader(status)
	fmt.Fprintf(
		ctx.Res,
		format,
		a...,
	)
}

// Writes the JSON response to the context's http response and responds with http.StatusOK
func (ctx *Context) RespondJson(response interface{}, status int) {
	json, err := json.Marshal(response)
	if err != nil {
		ctx.Error(err, http.StatusInternalServerError)
		return
	}
	if json, err = prettyprint(json); err != nil {
		ctx.Error(err, http.StatusInternalServerError)
		return
	}
	ctx.Res.Header().Set("Content-Type", "application/json")
	ctx.Respond(string(json), status)
}

// Writes the error as json to the context's http response and responds with the specified status
func (ctx *Context) Error(err error, status int) {
	ctx.RespondJson(entities.HttpError{
		Type: fmt.Sprintf("%T", err),
		Msg:  err.Error(),
	}, status)
}

func (ctx *Context) OpenWebsocket(station *entities.Station, isAdmin bool) {
	globalWebsockets.OpenWebsocket(ctx, station, isAdmin)
}


func (ctx *Context) CloseWebsocket() {
	globalWebsockets.CloseWebsocketCtx(ctx)
}

func (ctx *Context) Broadcast(station *entities.Station, header string, message string) {
	globalWebsockets.Broadcast(ctx, station, header, message)
}
