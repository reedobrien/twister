Twister
=======

Twister is a lightweight and modular framework for building web applications in
the [Go](http://golang.org/) programming language. Twister includes the
following features:

* Simple model for middleware using composition.
* Routing to handlers using regular expression match on host and path. 
* Cross site request forgery protection.
* Application specified error pages.
* Cookie parsing. 
* Static file handling. 
* Signed values for cookies and form parameters. 
* WebSockets. 
* Multipart forms. 
* Built-in HTTP server.
* OAuth 1.0 client
* and more...

Twister works great with templating engines and database drivers written for
Go. 

Installation
------------

Twister requires a working Go development environment. The 
[Getting Started](http://golang.org/doc/install.html) document
describes how to install the development environment. Once you have Go up and
running, you can install Twister with a single command:

    goinstall github.com/garyburd/twister/server

The Go distribution is Twister's only dependency. 
  
Documentation
-------------
 
* [web](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/web) - Defines the application interface to a server and includes functionality used by most web applications.
* [server](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/server) - An HTTP server impelemented in Go. 
* [oauth](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/oauth) - OAuth client. 
* [websocket](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/websocket) - WebSocket server implementation. 
* [expvar](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/expvar) - Exports variables as JSON over HTTP for monitoring. 
* [pprof](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/pprof) - Exports profiling data for the pprof tool.
* [gae](http://gopkgdoc.appspot.com/pkg/github.com/garyburd/twister/gae) - Support for running Twister on Google App Engine.

Examples
--------
 
* [wiki](http://github.com/garyburd/twister/tree/master/examples/wiki) - The [Go web application example](http://golang.org/doc/codelab/wiki/) converted to use Twister instead of the Go http package. 
* [demo](http://github.com/garyburd/twister/tree/master/examples/demo) -  Illustrates basic features of Twister. 
* [twitter](http://github.com/garyburd/twister/tree/master/examples/twitter) - Login to Twitter with OAuth and display home timeline. 
* [facebook](http://github.com/garyburd/twister/tree/master/examples/facebook) - Login to Facebook with OAuth2 and display news feed. 

License
-------

Twister is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

Discussion
----------
 
Discussion related to the use and development of Twister is held at the
[Twister Users](http://groups.google.com/group/twister-users) group.

You can also contact the author through [Github](https://github.com/inbox/new/garyburd).

About
-----
 
Twister was written by [Gary Burd](http://gary.beagledreams.com/). The name
"Twister" was inspired by [Tornado](http://tornadoweb.org/).
