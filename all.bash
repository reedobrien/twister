#!/usr/bin/env bash

for dir in web server websocket expvar pprof adapter examples/demo examples/chat examples/wiki
do
    (cd $dir && pwd && make DEPS= $*)
done
