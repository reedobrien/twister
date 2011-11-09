// Copyright 2010 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"bytes"
	"github.com/garyburd/twister/web"
	"log"
	"mime"
	"path"
	"text/template"
)

// itemFormatter formats a Facebook feed item by invoking the appropriate child
// template.
func itemFormatter(values ...interface{}) (interface{}, error) {
	itemType := values[0].(map[string]interface{})["type"].(string)
	t, ok := itemTemplates[itemType]
	if !ok {
		t = otherItemTemplate
	}
	var b bytes.Buffer
	err := t.t.Execute(&b, values[0])
	if err != nil {
		return nil, err
	}
	return b.String(), nil
}

type Template struct {
	t        *template.Template
	mimeType string
}

func parseTemplate(filename string) *Template {
	return &Template{
		t: template.Must(template.New(filename).
			Funcs(template.FuncMap{"item": itemFormatter}).
			ParseFile(path.Join("template", filename))),
		mimeType: mime.TypeByExtension(path.Ext(filename))}
}

func (t *Template) respond(req *web.Request, status int, value interface{}, kvs ...string) {
	header := web.NewHeader(kvs...)
	header.Set(web.HeaderContentType, t.mimeType)
	err := t.t.Execute(req.Responder.Respond(status, header), value)
	if err != nil {
		log.Println("Error executing template", err)
	}
}

var (
	homeTemplate      = parseTemplate("home.html")
	loggedOutTemplate = parseTemplate("loggedout.html")
	otherItemTemplate = parseTemplate("item/other.html")
	itemTemplates     = map[string]*Template{
		"link":   parseTemplate("item/link.html"),
		"status": parseTemplate("item/status.html"),
	}
)
