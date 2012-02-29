// Hello is a simple "hello world" application.
package main

import (
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	"io"
)

func serveHello(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain; charset=\"utf-8\"")
	io.WriteString(w, "Hello World!")
}

func main() {
	h := web.NewRouter().Register("/", "GET", serveHello)
	server.Run(":8080", h)
}
