#!/bin/bash
BUILD_VERSION=$(git describe --tags --always)
CID=$(git log --format="%H" -n 1)
if [ "$1" = "linux" ] || [ "$1" = "" ]; then
    echo "Generating Linux binary"
    env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.AppVersion=$BUILD_VERSION -X main.Sha=$CID" -o dist/slack-tube-service-$BUILD_VERSION-linux-amd64
fi
if [ "$1" = "windows" ] || [ "$1" = "" ]; then
    echo "Generating Windows binary"
    env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.AppVersion=$BUILD_VERSION -X main.Sha=$CID" -o dist/slack-tube-service-$BUILD_VERSION-windows-amd64.exe
fi
if [ "$1" = "mac" ] || [ "$1" = "" ]; then
    echo "Generating MacOS binary"
    env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.AppVersion=$BUILD_VERSION -X main.Sha=$CID" -o dist/slack-tube-service-$BUILD_VERSION-darwin-amd64
fi
echo "Done!"
ls -al dist/