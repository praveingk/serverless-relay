SW_VERSION ?= latest_tls
IMAGE_ORG ?= mcnet

IMAGE_TAG_BASE ?= quay.io/$(IMAGE_ORG)/client_function
IMG ?= $(IMAGE_TAG_BASE):$(SW_VERSION)
build:
	@echo "Start go build phase"
	go build -o ./bin/frelay ./cmd/frelay/frelay.go
	go build -o ./bin/fr-adm ./cmd/admin/admin.go
	go build -o ./bin/client_function ./client/client_function.go

docker-build:
	docker build --progress=plain --rm --tag client_function .

build-image:
	docker build --build-arg SW_VERSION="$(SW_VERSION)" -t ${IMG} .
push-image:
	docker push ${IMG}