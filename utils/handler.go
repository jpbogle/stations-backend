package utils

import (
	"log"
	"net/http"
	"time"
	"net"
	"bufio"
	"fmt"
)

var IS_DEBUG = false

type rootHandler struct {
	*Handler
}

func CreateRootHandler() *rootHandler {
	return &rootHandler{
		CreateHandler(),
	}
}

func (h *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	res := createStatusResponseWriter(w, IS_DEBUG)
	defer func() {
		log.Printf(
			"%s %s %s%v\x1b[0m %s\n",
			r.Method,
			r.URL.Path,
			res.StatusColor(),
			res.status,
			time.Since(start),
		)
	}()

	fields := make(map[string]string)
	handlerFunc := h.CallRoute(r.URL.Path, fields)
	if handlerFunc != nil {
		ctx := &Context{
			res,
			r,
			fields,
		}
		// log.Println(r.URL.Path)
		// if r.URL.Path == "/api/jbogle/station/ws" {
		// 	log.Println("WS")
		// 	go handlerFunc(ctx)
		// } else {
		// 	log.Println("NOT WS")
		// 	handlerFunc(ctx)
		// }
		handlerFunc(ctx)
	} else { //not found
		http.NotFound(res, r)
	}
}

//////////////////////////////////////////////////////////////////////////////
//
// StatusResponseWriter
//
//////////////////////////////////////////////////////////////////////////////

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}


func createStatusResponseWriter(w http.ResponseWriter, IS_DEBUG bool) *statusResponseWriter {
	if IS_DEBUG {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4000")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "http://stations.live")
	}
    w.Header().Set("Access-Control-Allow-Credentials", "true")
	return &statusResponseWriter{
		w,
		0,
	}
}

func (w *statusResponseWriter) StatusColor() string {
	prefix := w.status / 100
	switch prefix {
	case 0:
		return "\x1b[31m" // red
	case 1:
		return "\x1b[32m" // green
	case 2:
		return "\x1b[32m" // green
	case 3:
		return "\x1b[96m" // cyan
	case 4:
		return "\x1b[33m" // yellow
	case 5:
		return "\x1b[31m" // red
	default:
		return "\x1b[37m" // light gray
	}
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("not a Hijacker")
}

// func (w http.ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
//        return w.ResponseWriter.(http.Hijacker).Hijack()
// }

func SetDebug(debug bool) {
	IS_DEBUG = debug
}
