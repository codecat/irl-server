@echo off

set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o __deploy
scp __deploy irlpi:/home/nimble/irl/irl-server
scp config.toml irlpi:/home/nimble/irl/
del __deploy
