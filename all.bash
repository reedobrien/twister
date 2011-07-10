#!/usr/bin/env bash

for dir in . web server oauth websocket expvar pprof examples/demo examples/twitter examples/facebook examples/wiki
do
    (cd $dir; pwd; make DEPS= $*)
done
