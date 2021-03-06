#!/bin/bash

pre() {
    go vet ./...
    go fmt ./...
    /Users/rich_youngkin/Software/repos/go/bin/golint
    go test -race ./...
}

build() {
    cd src/cmd/todod
    go build
    cd -
}

buildARM() {
    cd src/cmd/todod
    /usr/bin/env GOOS=linux GOARCH=arm GOARM=7 go build
    cd -
}

buildImage() {
    doc
}

if [ $1 = "pre" ] 
then
    pre
elif [ $1 = "build" ] 
then
    build
elif [ $1 = "buildARM" ] 
then
    buildARM
else
    echo "usage:"
    echo "  build.sh [pre | build | buildARM | tag | all]"
    exit 1
fi