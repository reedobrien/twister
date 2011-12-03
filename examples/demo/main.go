package main

import (
	"errors"
	"flag"
	"github.com/garyburd/twister/adapter"
	"github.com/garyburd/twister/expvar"
	"github.com/garyburd/twister/pprof"
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	"log"
	"net"
	"net/http"
	"text/template"
)

func xsrf(req *web.Request) string {
	if req == nil {
		return ""
	}
	return req.Param.Get(web.XSRFParamName)
}

var templates = template.Must(template.New("set").Funcs(template.FuncMap{"xsrf": xsrf}).ParseGlob("*.html"))

func handler(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "home.html", map[string]interface{}{"req": req}); err != nil {
		log.Print(err)
	}
}

func errorHandler(req *web.Request, status int, reason error, header web.Header) {
	w := req.Responder.Respond(status, header)
	if err := templates.ExecuteTemplate(w, "home.html", map[string]interface{}{"req": req, "status": status, "reason": reason}); err != nil {
		log.Print(err)
	}
}

func panicBeforeResponse(req *web.Request) {
	panic(errors.New("Panic Attack!"))
	handler(req)
}

func panicAfterResponse(req *web.Request) {
	handler(req)
	panic(errors.New("Panic Attack!"))
}

func multipartHandler(req *web.Request) {
	files, err := web.ParseMultipartForm(req, -1)
	if err != nil {
		req.Error(web.StatusBadRequest, err)
		return
	}
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "home.html", map[string]interface{}{"req": req, "files": files}); err != nil {
		log.Print(err)
	}
}

var addr = flag.String("addr", ":8080", "http service address")
var useAdapter = flag.Bool("adapter", false, "use net/http adapter")

func main() {
	flag.Parse()
	r := web.NewRouter().
		Register("/", "*", web.FormHandler(-1, false, web.HandlerFunc(handler))).
		Register("/static/<path:.*>", "GET", web.DirectoryHandler("static/", nil)).
		Register("/example/file.txt", "GET", web.FileHandler("static/file.txt", nil)).
		Register("/urlparam/<a>/<b>", "GET", handler).
		Register("/panic/before", "GET", web.HandlerFunc(panicBeforeResponse)).
		Register("/panic/after", "GET", web.HandlerFunc(panicAfterResponse)).
		Register("/limit", "POST", web.FormHandler(16, false, web.HandlerFunc(handler))).
		Register("/xsrf", "*", web.FormHandler(-1, true, web.HandlerFunc(handler))).
		Register("/multipart", "POST", multipartHandler).
		Register("/debug/expvar", "GET", expvar.ServeWeb).
		Register("/debug/pprof/<:.*>", "*", pprof.ServeWeb).
		Register("/proxy", "GET", web.ProxyHeaderHandler("X-Real-Ip", "X-Scheme", web.HandlerFunc(handler)))

	h := web.SetErrorHandler(errorHandler, r)

	if *useAdapter {
		log.Print("Running with adapter.")
		if err := http.ListenAndServe(*addr, adapter.HTTPHandler{h}); err != nil {
			log.Fatal(err)
		}
	} else {
		listener, err := net.Listen("tcp", *addr)
		if err != nil {
			log.Fatal("Listen", err)
		}
		defer listener.Close()
		if err := (&server.Server{Listener: listener, Handler: h, Logger: server.LoggerFunc(server.VerboseLogger)}).Serve(); err != nil {
			log.Fatal(err)
		}
	}
}
