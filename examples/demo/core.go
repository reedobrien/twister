package main

import (
	"github.com/garyburd/twister/web"
	"os"
	"template"
)

func coreErrorHandler(req *web.Request, status int, reason os.Error, header web.Header) {

	coreTempl.Execute(
		req.Responder.Respond(status, header),
		map[string]interface{}{
			"req":     req,
			"status":  status,
			"message": reason,
			"xsrf":    req.Param.Get(web.XSRFParamName),
		})
}

func coreHandler(req *web.Request) {
	if req.Param.Get("panic") == "before" {
		panic(os.NewError("Panic Attack!"))
	}
	coreTempl.Execute(
		req.Respond(web.StatusOK, web.HeaderContentType, "text/html"),
		map[string]interface{}{
			"req":     req,
			"status":  web.StatusOK,
			"message": "ok",
			"xsrf":    req.Param.Get(web.XSRFParamName),
		})
	if req.Param.Get("panic") == "after" {
		panic(os.NewError("Panic Attack!"))
	}
}

var coreTempl = template.Must(template.New("core").Parse(coreStr))

const coreStr = `
<html>
<head>
<title>Core</title>
</head>
<body>
<hr>
Status: {{html .status}} {{html .message}}
<hr>
<a href="/core/file">file handler</a><br>
<a href="/static/file.txt">directory handler</a><br>
<a href="/core">/core</a><br>
<a href="/core/a/blorg">/core/a/blorg</a><br>
<a href="/core/a/foo?b=bar&amp;c=quux">/core/a/foo?b=bar&amp;c=quux</a><br>
<a href="/core/a/blorg/">/core/a/blorg/</a><br>
<a href="/core/b/foo/c/bar">/core/b/foo/c/bar</a><br> 
<a href="/core/b/foo/c/bar/">/core/b/foo/c/bar/</a> (not found)<br>
<a href="/core/?panic=before">/core/?panic=before</a><br>
<a href="/core/?panic=after">/core/?panic=after</a><br>
<form method="post" action="/core/c"><input type="hidden" name="xsrf" value="{{.xsrf}}"><input type=text value="hello" name=b><input type="submit"></form>
<form method="post" action="/core/c"><input type=text value="hello" name=b><input value="xsrf fail" type="submit"></form>
<hr>
{{with .req}}
<table>
<tr><th align="left" valign="top">RemoteAddr</th><td>{{html .RemoteAddr}}</td></tr>
<tr><th align="left" valign="top">Method</th><td>{{html .Method}}</td></tr>
<tr><th align="left" valign="top">URL</th><td>{{html .URL}}</td></tr>
<tr><th align="left" valign="top">ProtocolVersion</th><td>{{html .ProtocolVersion}}</td></tr>
<tr><th align="left" valign="top">Param</th><td>{{html .Param}}</td></tr>
<tr><th align="left" valign="top">URLParam</th><td>{{html .URLParam}}</td></tr>
<tr><th align="left" valign="top">ContentType</th><td>{{html .ContentType}}</td></tr>
<tr><th align="left" valign="top">ContentLength</th><td>{{html .ContentLength}}</td></tr>
<tr><th align="left" valign="top">Header</th><td>{{html .Header}}</td></tr>
</table>
{{end}}
</body>
</html> `
