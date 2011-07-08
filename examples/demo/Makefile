run: clean example
	./example 
 
include $(GOROOT)/src/Make.inc

TARG=example
DEPS=../../server ../../websocket ../../pprof ../../expvar
GOFILES=\
    main.go\
    core.go\
    multipart.go\

include $(GOROOT)/src/Make.cmd
