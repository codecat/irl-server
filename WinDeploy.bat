@echo off

set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o __deploy
scp __deploy 10.0.0.2:/home/pi/encoder-server/encoder-server
scp config.toml 10.0.0.2:/home/pi/encoder-server/
del __deploy
