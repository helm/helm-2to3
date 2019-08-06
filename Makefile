HELM_HOME ?= $(shell helm home)
HELM_PLUGIN_NAME := 2to3
HELM_PLUGIN_DIR ?= $(HELM_HOME)/plugins/helm-2to3
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)
DIST := $(CURDIR)/_dist
LDFLAGS := "-X main.version=${VERSION}"

PHONY: build
build:
	go build -o bin/${HELM_PLUGIN_NAME} -ldflags $(LDFLAGS) ./main.go
