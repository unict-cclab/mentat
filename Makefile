IMG ?= ghcr.io/unict-cclab/mentat:latest

run:
	go run main.go

build:
	go build

build-image:
	docker build -t ${IMG} .

push-image:
	docker push ${IMG}