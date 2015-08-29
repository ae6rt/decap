#!/bin/sh


docker --tlsverify=false run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp ae6rt/golang:1.5 make
