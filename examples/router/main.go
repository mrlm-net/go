package main

import (
	"fmt"
	"net/http"

	"github.com/mrlm-net/go/pkg/router"
)

func main() {
	srv := &Server{
		router: router.New[HTTPContext](),
	}

	// Clean! Just *HTTPContext
	srv.router.AddRoute("GET", "/users/:id", func(ctx *HTTPContext) {
		// Direct access to everything!
		fmt.Println("User ID:", ctx.Params[0].Value)
		fmt.Println("Request:", ctx.Request.URL.Path)
	})

	fmt.Println(srv)
}

type Server struct {
	router *router.Router[HTTPContext]
}

type HTTPContext struct {
	Request  *http.Request
	Response http.ResponseWriter
	Params   router.Params // Embed params directly
	// Add any other fields you need
	index int8
}
