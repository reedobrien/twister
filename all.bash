#!/usr/bin/env bash

for dir in web server oauth websocket expvar pprof adapter examples/demo examples/twitter examples/wiki
do
    (cd $dir; pwd; make DEPS= $*)
done
