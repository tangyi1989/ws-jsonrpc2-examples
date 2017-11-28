GO = go
ENV = export GOPATH=`pwd`/lib:`pwd`; export GOROOT=$(GOROOT);
GOGET = export GOPATH=`pwd`/lib; $(GO) get
BUILD = $(ENV)  $(GO) build

dep:
	mkdir -p `pwd`/bin
	mkdir -p `pwd`/lib/src
	$(GOGET) github.com/gorilla/websocket

run-chat: dep
	$(ENV) $(GO) run src/chat.go
