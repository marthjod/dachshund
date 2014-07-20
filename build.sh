#!/bin/bash

for p in errorcategory patternchecker spellchecker
do
    mkdir -p $GOPATH/src/$p && \
    go build $p.go && \
    cp $p.go $GOPATH/src/$p && \
    go install $p
done

go build -o dachshund main.go
