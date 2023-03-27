IMG_NAME ?= ghcr.io/amarchese96/mentat
IMG_TAG ?= latest

run:
	go run main.go

build:
	go build

docker-build:
	docker build -t ${IMG_NAME}:${IMG_TAG} .

docker-push:
	docker push ${IMG_NAME}:${IMG_TAG}
