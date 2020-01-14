.PHONY: build upload

build: parts/*.txt
	go run main.go > ads.txt

