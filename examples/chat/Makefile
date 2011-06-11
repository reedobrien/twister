run: clean chat
	./chat 
 
include $(GOROOT)/src/Make.inc

TARG=chat
DEPS=../../server ../../websocket ../../pprof
GOFILES=\
    chat.go

include $(GOROOT)/src/Make.cmd
