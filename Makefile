all:
	@echo "Usage: make <deploy|test>"

deploy:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o __deploy
	scp __deploy 10.0.0.2:/home/pi/encoder-server/encoder-server
	scp config.toml 10.0.0.2:/home/pi/encoder-server/
	rm __deploy

test: deploy
	ssh -t 10.0.0.2 "cd /home/pi/encoder-server; ./encoder-server"
