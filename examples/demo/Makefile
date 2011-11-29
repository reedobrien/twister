run: clean example
	./example 
 
run-adapter: clean example
	./example -adapter

include $(GOROOT)/src/Make.inc

TARG=example
DEPS=../../server ../../web ../../pprof ../../expvar ../../adapter
GOFILES=\
    main.go\

include $(GOROOT)/src/Make.cmd
