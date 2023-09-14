include VERSION.mk

CONTAINER_REPO := us-docker.pkg.dev/jarvice/images
BUILD_ARGS := --build-arg "VERSION=${VERSION}"
BUILD_ARGS += --build-arg "TARGETARCH=amd64"
BUILD_ARGS += --build-arg "BUILD=$(shell git rev-parse --short HEAD)-$(shell date -u '+%Y%m%d%H%M')"
BUILD_ARGS += --build-arg "DRAGEN_LIC=${DRAGEN_LIC}"
BUILD_ARGS += --label "maintainer=Nimbix"
BUILD_ARGS += --label "net.eviden.version=${VERSION}"
BUILD_ARGS += --label "net.eviden.commit-id=${shell git rev-parse --short HEAD}"
BUILD_ARGS += --label 'net.eviden.license_terms=Copyright (c) 2023 Nimbix, Inc.  All Rights Reserved.'

all: meter service

publish: all push-meter push-service

service:
	docker build --add-host "metadata.google.internal:127.0.0.1" -t "${CONTAINER_REPO}/jarvice-dragen-service:${VERSION}" --build-arg "PACKAGE=service" ${BUILD_ARGS} ${PWD}

push-service:
	docker push "${CONTAINER_REPO}/jarvice-dragen-service:${VERSION}"

meter:
	docker build --add-host "metadata.google.internal:127.0.0.1" -t "${CONTAINER_REPO}/jarvice-dragen-meter:${VERSION}" --build-arg "PACKAGE=meter" ${BUILD_ARGS} -f Dockerfile.slim ${PWD}

push-meter:
	docker push "${CONTAINER_REPO}/jarvice-dragen-meter:${VERSION}"
