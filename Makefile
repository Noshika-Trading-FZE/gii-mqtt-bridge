#
# Copyright: Pixel Networks <support@pixel-networks.com> 
# Author: Oleg Borodin <oleg.borodin@pixel-networks.com>
#

GO_CMD   = go
GO_ENV   = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 
IMAGE    = pix-mqtt-bridge
REGISTRY = pixelcore.azurecr.io

TARGET   = ./pmbri
VER = $(shell cat VERSION)

all:

build: ${TARGET}

${TARGET}: pmbri.go
	env ${GO_ENV} ${GO_CMD} build  -o $@ $<

image:
	sudo docker build . -t ${IMAGE} -t ${IMAGE}:${VER} -t ${IMAGE}:latest -t ${IMAGE}:staging -t ${REGISTRY}/${IMAGE}:latest -t ${REGISTRY}/${IMAGE}:${VER}

start:
	docker run ${REGISTRY}/${IMAGE}:latest

clean:
	rm -f ${TARGET}

version:
	git tag $(shell ./update-version)
	git ci -m 'version update' VERSION

push:
	git push --tags
	git push
#	docker push ${REGISTRY}/${IMAGE}:latest
#	docker push ${REGISTRY}/${IMAGE}:${VER}
#EOF
