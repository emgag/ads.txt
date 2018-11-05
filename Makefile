.PHONY: build upload

build: parts/*.txt
	go run main.go > ads.txt

upload: build
	rclone -v --config .rclone.conf copyto ads.txt ads:ads/ads.txt
