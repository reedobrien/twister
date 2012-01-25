#!/usr/bin/env bash
B=github.com/garyburd/twister
go $* $B/web $B/server $B/websocket $B/expvar $B/pprof $B/adapter $B/examples/demo $B/examples/chat $B/examples/wiki
