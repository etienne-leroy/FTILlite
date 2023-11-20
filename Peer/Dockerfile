# CUDA Version: 11.8.0
# Go Version: 1.18.3

# Step 1: Create a base stage that can be used for building the Go code. We do this first so that the Docker engine
# will (potentially) cache the final layer of this stage and use it for subsequent builds.
#
# Note: depending on how the CI environment is configured, the caching may not work and we will have to create a separate 
#       Docker image that is pushed to the registry and used to run the 'build-go' stage instead.
FROM ubuntu:22.04 AS go-ubuntu

RUN apt-get update
RUN apt-get install -y wget git gcc

RUN wget -P /tmp https://go.dev/dl/go1.18.3.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf /tmp/go1.18.3.linux-amd64.tar.gz
RUN rm /tmp/go1.18.3.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Step 2. Build the ftillite Go code using the previously built go-ubuntu layer

FROM go-ubuntu AS build-go
RUN mkdir -p /go/src/github.com/AUSTRAC/ftillite/Peer
WORKDIR /go/src/github.com/AUSTRAC/ftillite/Peer

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY segment ./segment
COPY lib/libftcrypto/src/ftcrypto.h ./lib/ftcrypto.h
COPY lib/libftcrypto.so /lib/libftcrypto.so

RUN cd ./segment && go test -v ./...
RUN go build -o ./ftillite-peer ./cmd/segment/main.go

# Step 3. Create the (final) image that will be used to run ftillite. 
FROM nvidia/cuda:11.8.0-base-ubuntu22.04 AS runtime
WORKDIR /app

COPY --from=build-go /go/src/github.com/AUSTRAC/ftillite/Peer/ftillite-peer /app/ftillite-peer
COPY --from=build-go /lib/libftcrypto.so /lib/libftcrypto.so

ENV FTILITE_ENABLE_GPU=false

CMD ["/app/ftillite-peer"]