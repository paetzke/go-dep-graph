#! /bin/bash

./go-dep-graph.amd64 "${@}" | dot -Tpdf -O
