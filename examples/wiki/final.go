// Wiki is a port of the Go wiki example to Twister.
// http://golang.org/doc/articles/wiki/
package main

import (
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	"io/ioutil"
	"log"
	"text/template"
)

type page struct {
	Title string
	Body  []byte
}

func (p *page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &page{Title: title, Body: body}, nil
}

func viewHandler(req *web.Request) {
	title := req.URLParam["title"]
	p, err := loadPage(title)
	if err != nil {
		req.Redirect("/edit/"+title, false)
		return
	}
	renderTemplate(req, "view", p)
}

func editHandler(req *web.Request) {
	title := req.URLParam["title"]
	p, err := loadPage(title)
	if err != nil {
		p = &page{Title: req.URLParam["title"]}
	}
	renderTemplate(req, "edit", p)
}

func saveHandler(req *web.Request) {
	body := req.Param.Get("body")
	title := req.URLParam["title"]
	p := &page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		req.Error(web.StatusInternalServerError, err)
		return
	}
	req.Redirect("/view/"+title, false)
}

var templates = make(map[string]*template.Template)

func init() {
	for _, tmpl := range []string{"edit", "view"} {
		t := template.Must(template.ParseFiles(tmpl + ".html"))
		templates[tmpl] = t
	}
}

func renderTemplate(req *web.Request, tmpl string, p *page) {
	err := templates[tmpl].Execute(
		req.Respond(web.StatusOK,
					web.HeaderContentType, "text/html; charset=\"utf-8\""),
		map[string]interface{}{
			"page": p,
			"xsrf": req.Param.Get("xsrf"),
		})
	if err != nil {
		log.Println("error rendering", tmpl, err)
	}
}

func main() {
	const titleParam = "<title:[a-zA-Z0-9]+>"
	h := web.FormHandler(10000, true, // limit size of form to 10k, enable xsrf
		web.NewRouter().
			Register("/view/"+titleParam, "GET", viewHandler).
			Register("/edit/"+titleParam, "GET", editHandler, "POST", saveHandler))
	server.Run(":8080", h)
}
