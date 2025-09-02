build:
	pwd
	go build cmd/main.go

run:
	./main

include .env
buildForLinux:
	pwd
	go build cmd/main.go

