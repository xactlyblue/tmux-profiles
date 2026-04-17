.PHONY: build clean install

clean: 
	rm -rf bin

build: main.go
	go build -o bin/tmux-profiles main.go

install: build
	mv bin/tmux-profiles /usr/local/bin
