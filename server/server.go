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

// The server package implements the HTTP protocol for a web server.
package server

import (
	"bufio"
	"bytes"
	"github.com/garyburd/twister/web"
	"http"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

var (
	ErrBadRequestLine = os.NewError("could not parse request line")
)

// Server defines parameters for running an HTTP server.
type Server struct {
	// The server accepts incoming connections on this listener. The
	// application is required to set this field.
	Listener net.Listener

	// The server dispatches requests to this handler. The application is
	// required to set this field.
	Handler web.Handler

	// If true, then set the request URL protocol to HTTPS.
	Secure bool

	// Set request URL host to this string if host is not specified in the
	// request or headers.
	DefaultHost string

	// The net.Conn.SetReadTimeout value for new connections.
	ReadTimeout int64

	// The net.Conn.SetWriteTimeout value for new connections.
	WriteTimeout int64

	// Log the request.
	Logger Logger

	// If true, do not recover from handler panics.
	NoRecoverHandlers bool
}

// Logger defines an interface for logging a request.
type Logger interface {
	Log(lr *LogRecord)
}

// LoggerFunc is a type adapter to allow the use of ordinary functions as Logger.
type LoggerFunc func(*LogRecord)

// Log calls f(lr).
func (f LoggerFunc) Log(lr *LogRecord) { f(lr) }

// transaction represents a single request-response transaction.
type transaction struct {
	server             *Server
	conn               net.Conn
	br                 *bufio.Reader
	responseBody       responseBody
	chunked            bool
	closeAfterResponse bool
	hijacked           bool
	req                *web.Request
	requestAvail       int
	requestErr         os.Error
	respondCalled      bool
	responseErr        os.Error
	write100Continue   bool
	status             int
	header             web.HeaderMap
	headerSize         int
}

var requestLineRegexp = regexp.MustCompile("^([_A-Za-z0-9]+) ([^ ]+) HTTP/([0-9]+)\\.([0-9]+)[\r\n ]+$")

func readRequestLine(b *bufio.Reader) (method string, url string, version int, err os.Error) {

	p, err := b.ReadSlice('\n')
	if err != nil {
		if err == bufio.ErrBufferFull {
			err = web.ErrLineTooLong
		}
		return
	}

	m := requestLineRegexp.FindSubmatch(p)
	if m == nil {
		err = ErrBadRequestLine
		return
	}

	method = string(m[1])

	major, err := strconv.Atoi(string(m[3]))
	if err != nil {
		return
	}

	minor, err := strconv.Atoi(string(m[4]))
	if err != nil {
		return
	}

	version = web.ProtocolVersion(major, minor)

	url = string(m[2])

	return
}

func (t *transaction) prepare() (err os.Error) {
	method, rawURL, version, err := readRequestLine(t.br)
	if err != nil {
		return err
	}

	header := web.HeaderMap{}
	err = header.ParseHttpHeader(t.br)
	if err != nil {
		return err
	}

	url, err := http.ParseURL(rawURL)
	if err != nil {
		return err
	}

	if url.Host == "" {
		url.Host = header.Get(web.HeaderHost)
		if url.Host == "" {
			url.Host = t.server.DefaultHost
		}
	}

	if t.server.Secure {
		url.Scheme = "https"
	} else {
		url.Scheme = "http"
	}

	req, err := web.NewRequest(t.conn.RemoteAddr().String(), method, url, version, header)
	if err != nil {
		return
	}
	t.req = req

	t.requestAvail = req.ContentLength
	if t.requestAvail < 0 {
		t.requestAvail = 0
	}

	if s := req.Header.Get(web.HeaderExpect); s != "" {
		t.write100Continue = strings.ToLower(s) == "100-continue"
	}

	connection := strings.ToLower(req.Header.Get(web.HeaderConnection))
	if version >= web.ProtocolVersion(1, 1) {
		t.closeAfterResponse = connection == "close"
	} else if version == web.ProtocolVersion(1, 0) && req.ContentLength >= 0 {
		t.closeAfterResponse = connection != "keep-alive"
	} else {
		t.closeAfterResponse = true
	}

	req.Responder = t
	req.Body = requestReader{t}
	return nil
}

type requestReader struct {
	*transaction
}

func (t requestReader) Read(p []byte) (int, os.Error) {
	if t.requestErr != nil {
		if t.requestErr == web.ErrInvalidState {
			log.Println("twister: Request Read after response started.")
		}
		return 0, t.requestErr
	}
	if t.write100Continue {
		t.write100Continue = false
		io.WriteString(t.conn, "HTTP/1.1 100 Continue\r\n\r\n")
	}
	if t.requestAvail <= 0 {
		t.requestErr = os.EOF
		return 0, t.requestErr
	}
	if len(p) > t.requestAvail {
		p = p[0:t.requestAvail]
	}
	var n int
	n, t.requestErr = t.br.Read(p)
	t.requestAvail -= n
	return n, t.requestErr
}

func (t *transaction) Respond(status int, header web.HeaderMap) (body web.ResponseBody) {
	if t.hijacked {
		log.Println("twister: Respond called on hijacked connection")
		return nil
	}
	if t.respondCalled {
		log.Println("twister: Multiple calls to Respond")
		return nil
	}
	t.respondCalled = true
	t.requestErr = web.ErrInvalidState
	t.status = status
	t.header = header

	if te := header.Get(web.HeaderTransferEncoding); te != "" {
		log.Println("twister: transfer encoding not allowed")
		header[web.HeaderTransferEncoding] = nil, false
	}

	if t.requestAvail > 0 {
		t.closeAfterResponse = true
	}

	t.chunked = true
	contentLength := -1

	if status == web.StatusNotModified {
		header[web.HeaderContentType] = nil, false
		header[web.HeaderContentLength] = nil, false
		t.chunked = false
	} else if s := header.Get(web.HeaderContentLength); s != "" {
		contentLength, _ = strconv.Atoi(s)
		t.chunked = false
	} else if t.req.ProtocolVersion < web.ProtocolVersion(1, 1) {
		t.closeAfterResponse = true
	}

	if t.closeAfterResponse {
		header.Set(web.HeaderConnection, "close")
		t.chunked = false
	}

	if t.req.Method == "HEAD" {
		t.chunked = false
	}

	if t.chunked {
		header.Set(web.HeaderTransferEncoding, "chunked")
	}

	proto := "HTTP/1.0"
	if t.req.ProtocolVersion >= web.ProtocolVersion(1, 1) {
		proto = "HTTP/1.1"
	}
	statusString := strconv.Itoa(status)
	text := web.StatusText(status)

	var b bytes.Buffer
	b.WriteString(proto)
	b.WriteString(" ")
	b.WriteString(statusString)
	b.WriteString(" ")
	b.WriteString(text)
	b.WriteString("\r\n")
	header.WriteHttpHeader(&b)
	t.headerSize = b.Len()

	const bufferSize = 4096
	switch {
	case t.req.Method == "HEAD":
		t.responseBody, _ = newNullResponseBody(t.conn, b.Bytes())
	case t.chunked:
		t.responseBody, _ = newChunkedResponseBody(t.conn, b.Bytes(), bufferSize)
	default:
		t.responseBody, _ = newIdentityResponseBody(t.conn, b.Bytes(), bufferSize, contentLength)
	}
	return t.responseBody
}

func (t *transaction) Hijack() (conn net.Conn, buf []byte, err os.Error) {
	if t.respondCalled {
		return nil, nil, web.ErrInvalidState
	}

	conn = t.conn
	buf, err = t.br.Peek(t.br.Buffered())
	if err != nil {
		panic("twister.server: unexpected error peeking at bufio")
	}

	if t.server.Logger != nil {
		t.server.Logger.Log(&LogRecord{
			Request:  t.req,
			Header:   t.header,
			Hijacked: true,
		})
	}

	t.hijacked = true
	t.requestErr = web.ErrInvalidState
	t.responseErr = web.ErrInvalidState
	t.req = nil
	t.br = nil
	t.conn = nil

	return
}

// Finish the HTTP request
func (t *transaction) finish() os.Error {
	if !t.respondCalled {
		t.req.Respond(web.StatusOK, web.HeaderContentType, "text/html charset=utf-8")
	}
	var written int
	if t.responseErr == nil {
		written, t.responseErr = t.responseBody.finish()
	}
	if t.responseErr != nil {
		t.closeAfterResponse = true
	} else {
		t.responseErr = web.ErrInvalidState
	}
	if t.server.Logger != nil {
		err := t.responseErr
		if err == web.ErrInvalidState {
			err = t.requestErr
			if err == web.ErrInvalidState {
				err = nil
			}
		}
		t.server.Logger.Log(&LogRecord{
			Written:    written,
			Request:    t.req,
			Header:     t.header,
			HeaderSize: t.headerSize,
			Status:     t.status,
			Error:      err})
	}
	t.conn = nil
	t.br = nil
	t.responseBody = nil
	return nil
}

func (s *Server) serveConnection(conn net.Conn) {
	if s.ReadTimeout != 0 {
		conn.SetReadTimeout(s.ReadTimeout)
	}
	if s.WriteTimeout != 0 {
		conn.SetWriteTimeout(s.WriteTimeout)
	}
	br := bufio.NewReader(conn)
	for {
		t := transaction{
			server: s,
			conn:   conn,
			br:     br}
		if err := t.prepare(); err != nil {
			if err != os.EOF {
				log.Println("twister/server: prepare failed", err)
			}
			break
		}

		defer func() {
			if !s.NoRecoverHandlers {
				if r := recover(); r != nil {
					url := t.req.URL.String()
					stack := string(debug.Stack())
					log.Printf("Panic while serving \"%s\": %v\n%s", url, r, stack)
				}
			}
		}()

		s.Handler.ServeWeb(t.req)
		if t.hijacked {
			return
		}
		if err := t.finish(); err != nil {
			log.Println("twister/server: finish failed", err)
			break
		}
		if t.closeAfterResponse {
			break
		}
	}
	conn.Close()
}

// Serve accepts incoming HTTP connections on s.Listener, creating a new
// goroutine for each. The goroutines read requests and then call s.Handler to
// respond to the request.
//
// The "Hello World" server using Serve() is:
//
//  package main
//  
//  import (
//      "github.com/garyburd/twister/web"
//      "github.com/garyburd/twister/server"
//      "io"
//      "log"
//      "net"
//  )
//
//  func helloHandler(req *web.Request) {
//      w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain")
//      io.WriteString(w, "Hello, World!\n")
//  }
//  
//  func main() {
//      handler := web.NewRouter().Register("/", "GET", helloHandler)
//      listener, err := net.Listen("tcp", ":8080")
//      if err != nil {
//          log.Fatal("Listen", err)
//      }
//      defer listener.Close()
//      err = (&server.Server{Listener: listener, Handler: handler}).Serve()
//      if err != nil {
//          log.Fatal("Server", err)
//      }
//  }
func (s *Server) Serve() os.Error {
	for {
		conn, e := s.Listener.Accept()
		if e != nil {
			return e
		}
		go s.serveConnection(conn)
	}
	return nil
}

// Run is a convenience function for running an HTTP server. Run listens on the
// TCP address addr, initializes a server object and calls the server's Serve()
// method to handle HTTP requests. Run logs a fatal error if it encounters an
// error.
//
// The Server object is initialized with the handler argument and listener. If
// the application needs to set any other Server fields or if the application
// needs to create the listener, then the application should directly create
// the Server object and call the Serve() method.
//
// The "Hello World" server using Run() is:
//
//  package main
//  
//  import (
//      "github.com/garyburd/twister/web"
//      "github.com/garyburd/twister/server"
//      "io"
//  )
//  
//  func helloHandler(req *web.Request) {
//      w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain")
//      io.WriteString(w, "Hello, World!\n")
//  }
//  
//  func main() {
//      server.Run(":8080", web.NewRouter().Register("/", "GET", helloHandler))
//  }
//
func Run(addr string, handler web.Handler) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listen", err)
		return
	}
	defer listener.Close()
	err = (&Server{Logger: LoggerFunc(ShortLogger), Listener: listener, Handler: handler}).Serve()
	if err != nil {
		log.Fatal("Server", err)
	}
}
