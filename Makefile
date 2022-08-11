BUILD_DIR = build
SERVICES = users things http writer reader authn mqtt
DOCKERS = $(addprefix docker_,$(SERVICES))
CGO_ENABLED ?= 0
GOARCH ?= amd64

define compile_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -mod=vendor -ldflags "-s -w" -o ${BUILD_DIR}/alpha-$(1) cmd/$(1)/main.go
endef

define make_docker
	$(eval svc=$(subst docker_,,$(1)))

	docker build \
		--no-cache \
		--build-arg SVC=$(svc) \
		--build-arg GOARCH=$(GOARCH) \
		--build-arg GOARM=$(GOARM) \
		--tag=alpha/$(svc) \
		-f docker/Dockerfile .
endef

all: $(SERVICES)

.PHONY: all $(SERVICES) dockers

clean:
	rm -rf ${BUILD_DIR}

install:
	cp ${BUILD_DIR}/* $(GOBIN)

proto:
	sudo protoc --gofast_out=plugins=grpc:. *.proto
	sudo protoc --gofast_out=plugins=grpc:. messaging/*.proto

$(SERVICES):
	$(call compile_service,$(@))

$(DOCKERS):
	$(call make_docker,$(@),$(GOARCH))

dockers: $(DOCKERS)

run:
	docker-compose -f docker/docker-compose.yml up
