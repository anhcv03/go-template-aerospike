### -----------------------
# --- Make variables
### -----------------------

# only evaluated if required by a recipe
# http://make.mad-scientist.net/deferred-simple-variable-expansion/

# go module name (as in go.mod)
GO_MODULE_NAME = $(eval GO_MODULE_NAME := $$(shell \
	(mkdir -p tmp data 2> /dev/null && cat .modulename 2> /dev/null) \
	|| (gsdev modulename 2> /dev/null | tee .modulename) || echo "unknown" \
))$(GO_MODULE_NAME)

# Common infos
MODULE_NAME := $(eval MODULE_NAME := $(shell echo $(GO_MODULE_NAME) | awk -F'/' '{print $$6}'))$(MODULE_NAME)
MODULE_VERSION := $(shell git describe --tags --always)
# https://medium.com/the-go-journey/adding-version-information-to-go-binaries-e1b79878f6f2
ARG_COMMIT = $(eval ARG_COMMIT := $$(shell \
	(git rev-list -1 HEAD 2> /dev/null) \
	|| (echo "unknown") \
))$(ARG_COMMIT)

ARG_BUILD_DATE = $(eval ARG_BUILD_DATE := $$(shell \
	(date -Is 2> /dev/null || date 2> /dev/null || echo "unknown") \
))$(ARG_BUILD_DATE)

# https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
LDFLAGS = $(eval LDFLAGS := "\
-X '$(GO_MODULE_NAME)/internal/config.ModuleName=$(MODULE_NAME)'\
-X '$(GO_MODULE_NAME)/internal/config.Commit=$(ARG_COMMIT)'\
-X '$(GO_MODULE_NAME)/internal/config.BuildDate=$(ARG_BUILD_DATE)'\
")$(LDFLAGS)

### -----------------------
# --- Building
### -----------------------

CGO_ENABLED ?= 0

go-build: ##- (opt) Runs go build.
	CGO_ENABLED=$(CGO_ENABLED) go build -mod=vendor -ldflags $(LDFLAGS) -o bin/${MODULE_NAME}

proto:
	rm -rf pkg/pb/*.go pkg/pb/ipc/*.go
	protoc --proto_path=/usr/local/include --proto_path=. --go_out=. --go-grpc_out=. pkg/proto/*.proto --experimental_allow_proto3_optional
	protoc --proto_path=/usr/local/include --proto_path=./pkg/proto/ipc-messages/ --go_out=./pkg/pb/ pkg/proto/ipc-messages/*.proto \
		pkg/proto/ipc-messages/ipc/track/tag/*.proto \
		pkg/proto/ipc-messages/ipc/track/*.proto \
		pkg/proto/ipc-messages/ipc/track/cmd/*.proto --experimental_allow_proto3_optional
	protoc --proto_path=/usr/local/include --proto_path=. --go_out=./pkg/pb/ --go-grpc_out=./pkg/pb/ pkg/proto/ag_algo/*.proto --experimental_allow_proto3_optional


#proto:
#	rm -rf pkg/pb/*.go pkg/pb/ipc/*.go
#	protoc --proto_path=/usr/local/include --proto_path=. --go_out=. --go-grpc_out=. pkg/proto/*.proto
#	protoc --proto_path=/usr/local/include --proto_path=./pkg/proto/ipc-messages/ --go_out=./pkg/pb/ pkg/proto/ipc-messages/*.proto \
		pkg/proto/ipc-messages/ipc/track/tag/*.proto \
		pkg/proto/ipc-messages/ipc/track/*.proto \
		pkg/proto/ipc-messages/ipc/track/cmd/*.proto
#	protoc --proto_path=/usr/local/include --proto_path=. --go_out=./pkg/pb/ --go-grpc_out=./pkg/pb/ pkg/proto/ag_algo/*.proto

# proto:
# 	@echo "running proto..."

# 	@rm -rf pkg/pb/*.go

# 	@protoc --proto_path=/usr/local/include --proto_path=. --go_out=. --go-grpc_out=. pkg/proto/*.proto pkg/proto/ag_algo/*.proto  pkg/proto/ipc-messages/ipc/track/cmd/*.proto pkg/proto/ipc-messages/ipc/track/tag/*.proto pkg/proto/ipc-messages/other/*.proto --experimental_allow_proto3_optional

# 	@cd pkg && ./custom.generate.sh

tidy: ##- (opt) Tidy our go.sum file.
	go mod tidy

vendor:
	go mod vendor

### -----------------------
# --- Helpers
### -----------------------

clean: ##- Cleans ./tmp and ./api/tmp folder.
	@echo "make clean"
	@rm -rf tmp 2> /dev/null
	@rm -rf api/tmp 2> /dev/null

get-module-name: ##- Prints current go module-name (pipeable).
	@echo "${GO_MODULE_NAME}"

info-module-name: ##- (opt) Prints current go module-name.
	@echo "go module-name: '${GO_MODULE_NAME}'"

set-module-name: ##- Wizard to set a new go module-name.
	@rm -rf .modulename
	@echo "Enter new go module-name:" \
		&& read new_module_name \
		&& echo "new go module-name: '$${new_module_name}'" \
		&& echo -n "Are you sure? [y/N]" \
		&& read ans && [ $${ans:-N} = y ] \
		&& echo -n "Please wait..." \
		&& find . -not -path '*/\.*' -not -path './Makefile' -type f -exec sed -i "s|${GO_MODULE_NAME}|gitlab.vht.vn/tt-kttt/lae-project/utm/simulator/$${new_module_name}|g" {} \; \
		&& echo "gitlab.vht.vn/tt-kttt/lae-project/utm/simulator/$${new_module_name}" >> .modulename \
		&& echo "new go module-name: '$${new_module_name}'!"

get-go-ldflags: ##- (opt) Prints used -ldflags as evaluated in Makefile used in make go-build
	@echo $(LDFLAGS)

tools: ##- (opt) Install packages as specified in tools.go.
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -P $$(nproc) -L 1 -tI % go install %

# debian package info
PKG_CFG=/etc/
PKG_SVC=/lib/systemd/system/
PKG_NAME= $(MODULE_NAME)
PKG_BIN=/usr/local/bin/$(PKG_NAME)/
PKG_VERSION= $(shell git describe --tags)
PKG_AUTH = $(shell git log -1 --pretty=format:"%ae")
PKG_DESP = "Service for managing $(MODULE_NAME)"
PKG_ARC=amd64
PKG_FOLDER = $(PKG_NAME)-$(PKG_ARC)-$(PKG_VERSION)

# Docker infos
DOCKER_REGISTRY ?= harbor.vht.vn
DOCKER_PROJECT ?= kttt
DOCKER_USER ?= admin
DOCKER_PASSWORD ?= 1
DOCKER_IMAGE_NAME ?= ${MODULE_NAME}
DOCKER_IMAGE_VERSION ?= ${MODULE_VERSION}
DOCKER_IMAGE := ${DOCKER_REGISTRY}/${DOCKER_PROJECT}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_VERSION}

# Service infos
DATA_DIR ?= data
STYLE_TIMEOUT ?= 1080000
MAX_STYLE_CONCURRENCY ?= 16

# Golang build infos
CGO_ENABLED ?= 0

### -----------------------
# --- Start dockers
### -----------------------

login-docker-registry:
	@echo 'logining docker to registry "${DOCKER_REGISTRY}"...'

	docker login \
		${DOCKER_REGISTRY} \
		-u ${DOCKER_USER} \
		-p ${DOCKER_PASSWORD}

build-docker-image:
	@echo 'building docker image "${DOCKER_IMAGE}"...'

	docker build \
		-t ${DOCKER_IMAGE} \
		-f ./Dockerfile .

push-docker-image:
	@echo 'pushing docker image "${DOCKER_IMAGE}"...'

	docker push \
		${DOCKER_IMAGE}

clean-docker-image:
	@echo 'cleaning docker image "${DOCKER_IMAGE}"...'

	docker rmi -f \
		${DOCKER_IMAGE}

release-docker-image:
	@echo 'releasing docker image "${DOCKER_IMAGE}"...'

	@make build-docker-image
	@make push-docker-image
	@make clean-docker-image

### -----------------------
# --- End dockers
### -----------------------


debian:
	make go-build
	@echo "make debian package"

	make debian-folder

	@cp deb/app.yaml deb/$(PKG_FOLDER)$(PKG_CFG)$(PKG_NAME)/app.yaml
	@mv bin/$(PKG_NAME) deb/$(PKG_FOLDER)$(PKG_BIN)/$(PKG_NAME)

	make debian-control
	make debian-service

	@dpkg-deb --build --root-owner-group deb/$(PKG_FOLDER)
	@curl -u "chdk:123456aA@" -H "Content-Type: multipart/form-data" --data-binary "@./deb/$(PKG_FOLDER).deb" "http://172.31.252.188:8081/repository/chdk-apt-hosted/"

debian-folder:
	$(eval PKG_FOLDER = $(PKG_NAME)-$(PKG_ARC)-$(PKG_VERSION))
	@mkdir -p deb/$(PKG_FOLDER)/DEBIAN
	@mkdir -p deb/$(PKG_FOLDER)$(PKG_BIN)
	@mkdir -p deb/$(PKG_FOLDER)$(PKG_BIN)/logs
	@mkdir -p deb/$(PKG_FOLDER)$(PKG_SVC)
	@mkdir -p deb/$(PKG_FOLDER)$(PKG_CFG)$(PKG_NAME)

debian-service:
	$(eval PKG_FOLDER = $(PKG_NAME)-$(PKG_ARC)-$(PKG_VERSION))
	$(eval svc_file = deb/$(PKG_FOLDER)$(PKG_SVC)$(PKG_NAME).service)
	@echo $(svc_file)
	@cp deb/service-template.service $(svc_file)
	sed -i 's|SVC_DESCRIPTION|${PKG_DESP}|g' $(svc_file)
	sed -i "s|SVC_BINARY|${PKG_BIN}${PKG_NAME}|g" $(svc_file)
	sed -i "s|SVC_CFG|${PKG_CFG}${PKG_NAME}|g" $(svc_file)
	sed -i "s|SVC_DIR|${PKG_BIN}|g" $(svc_file)

debian-control:
	@echo "Package: $(PKG_NAME)" > deb/$(PKG_FOLDER)/DEBIAN/control
	@echo "Version: $(PKG_VERSION)" >> deb/$(PKG_FOLDER)/DEBIAN/control
	@echo "Architecture: $(PKG_ARC)" >> deb/$(PKG_FOLDER)/DEBIAN/control
	@echo "Maintainer: $(PKG_AUTH)" >> deb/$(PKG_FOLDER)/DEBIAN/control
	@echo "Description: $(PKG_DESP)" >> deb/$(PKG_FOLDER)/DEBIAN/control

swag:
	swag init --parseDependency  --parseInternal --parseDepth 100  -g main.go
	swag fmt

server:
	@make go-build && ./bin/$(MODULE_NAME) start

.PHONY: tools vendor proto
