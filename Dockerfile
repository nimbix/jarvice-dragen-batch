############################################
# build stage
############################################
FROM us-docker.pkg.dev/jarvice/images/golang:1.21.1-alpine

# golang image uses /go by default, sets GOPATH
WORKDIR /go/src/jarvice.io/dragen
COPY . .
ARG GOROOT=/usr/local/go

ARG TARGETARCH
ARG GOARCH=${TARGETARCH}
ARG PACKAGE
ARG VERSION
ARG BUILD
ARG DRAGEN_LIC

RUN go get /go/src/jarvice.io/dragen/internal/jobs && \
	go get /go/src/jarvice.io/dragen/internal/google && \
	go get /go/src/jarvice.io/dragen/internal/monitor && \
	go get /go/src/jarvice.io/dragen/internal/logger && \
	go get /go/src/jarvice.io/dragen/cmd/${PACKAGE}

RUN CGO_ENABLED=0 GOOS=linux go build -o ${PACKAGE}.out -a \
	-ldflags "-X jarvice.io/dragen/config.Version=${VERSION} \
	-X jarvice.io/dragen/config.Build=${BUILD} \
	-X jarvice.io/dragen/config.DragenLic=${DRAGEN_LIC} \
	-extldflags -static -s -w" ./cmd/"${PACKAGE}"
# RUN test -n "${PACKAGE}" && gofmt -d -e ${PACKAGE} && CGO_ENABLED=0 GOOS=linux go build -o ${PACKAGE}.out -a -ldflags '-extldflags -static -s -w' ./"${PACKAGE}"

RUN mv ${PACKAGE}.out /usr/local/bin/entrypoint

############################################
# run stage
############################################
FROM google/cloud-sdk:440.0.0-slim

COPY --from=0 "/usr/local/bin/entrypoint" "/usr/local/bin/entrypoint"

ENTRYPOINT ["/usr/local/bin/entrypoint"]
