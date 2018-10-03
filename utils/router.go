package utils

import (
	"fmt"
	"net/http"
	"regexp"
	// "stations/entities"
)

// TODO: make routerStub a mapping of http method (GET, POST, PUT, DELETE) to function

// a router is either a handler or routerStub
type router interface {
	// Calls the specified route from the router's subroutes; returns nil if no route found, or
	// the matching subrouter
	CallRoute(path string, fields map[string]string) handlerFunc
}
type Handler struct { // *handler implements route
	verbatimSubroutes    []*subroute
	placeholderSubroutes []*subroute
}
type routerStub struct { // *routerStub implements route
	handlerFunc handlerFunc
}

type subroute struct {
	pattern   *regexp.Regexp
	subrouter router
}
type handlerFunc func(*Context)

//////////////////////////////////////////////////////////////////////////////
//
// Creators
//
//////////////////////////////////////////////////////////////////////////////

func CreateHandler() *Handler {
	return &Handler{
		make([]*subroute, 0),
		make([]*subroute, 0),
	}
}

func createHandlerFunc(handlerFunc func(*Context)) router {
	return &routerStub{
		handlerFunc,
	}
}

func createMethodWrapper(method string, f handlerFunc) router {
	wrapper := createHandlerFunc(func(ctx *Context) {
		if ctx.Req.Method == method {
			f(ctx)
		}
	})
	return wrapper
}

//////////////////////////////////////////////////////////////////////////////
//
// router methods
//
//////////////////////////////////////////////////////////////////////////////

func (r *Handler) CallRoute(path string, fields map[string]string) handlerFunc {
	checkSubroute := func(route *subroute) handlerFunc {
		// check if pattern matches subpath in order of router registration
		match := route.pattern.FindStringSubmatch(path)
		if match = route.pattern.FindStringSubmatch(path); match != nil {
			for i, name := range route.pattern.SubexpNames() {
				if i != 0 {
					fields[name] = match[i]
				}
			}
			// recursively call next subrouter
			subpath := fields["subpath"]
			delete(fields, "subpath")
			if stub := route.subrouter.CallRoute(subpath, fields); stub != nil {
				return stub
			}
		}
		return nil
	}
	// check routes with no placeholders for perfect match first
	for _, route := range r.verbatimSubroutes {
		if res := checkSubroute(route); res != nil {
			return res
		}
	}
	for _, route := range r.placeholderSubroutes {
		if res := checkSubroute(route); res != nil {
			return res
		}
	}
	return nil
}

// Add a method stub
func (r *Handler) addRouter(path string, subrouter router) {
	// replace ":paramname" with (?P<paramname>[^/]+) for regex matching
	placeholderRegexp := regexp.MustCompile(":[A-Za-z]+")
	hasPlaceholders := placeholderRegexp.Match([]byte(path))

	if hasPlaceholders {
		prefix := []byte("(?P<")
		suffix := []byte(">[^/]+)")
		path = string(placeholderRegexp.ReplaceAllFunc([]byte(path), func(match []byte) []byte {
			return append(append(prefix, match[1:]...), suffix...)
		}))
	}

	pattern := regexp.MustCompile(
		fmt.Sprintf("^%s(?P<subpath>/?.*)$", path), //ensures subpath is parsed in CallRoute
	)
	if hasPlaceholders {
		r.placeholderSubroutes = append(r.placeholderSubroutes, &subroute{pattern, subrouter})
	} else {
		r.verbatimSubroutes = append(r.verbatimSubroutes, &subroute{pattern, subrouter})
	}
}

// Add a handler
func (r *Handler) AddHandler(path string, handler router) {
	r.addRouter(path, handler)
}

func (r *Handler) Get(path string, f handlerFunc) {
	wrapper := createMethodWrapper(http.MethodGet, f)
	r.addRouter(path, wrapper)
}

func (r *Handler) Post(path string, f handlerFunc) {
	wrapper := createMethodWrapper(http.MethodPost, f)
	r.addRouter(path, wrapper)
}

func (r *Handler) Put(path string, f handlerFunc) {
	wrapper := createMethodWrapper(http.MethodPut, f)
	r.addRouter(path, wrapper)
}

func (r *Handler) Delete(path string, f handlerFunc) {
	wrapper := createMethodWrapper(http.MethodDelete, f)
	r.addRouter(path, wrapper)
}

//////////////////////////////////////////////////////////////////////////////
//
// routerStub methods
//
//////////////////////////////////////////////////////////////////////////////

func (r *routerStub) CallRoute(path string, fields map[string]string) handlerFunc {
	if len(path) != 0 && path != "/" {
		return nil
	}
	return r.handlerFunc
}
