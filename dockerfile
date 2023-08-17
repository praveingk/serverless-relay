FROM golang:1.19

# Create dockerfile with multi-stagets: stage 0: compile src and client
# Set destination for COPY
WORKDIR /client

# Copy the source code.
COPY . ./

# Build Go model
RUN CGO_ENABLED=0 go build -o ./bin/client_function ./tests/client_function.go
# Create dockerfile with multi-stagets :stage 1: low resources


FROM alpine:3.18

WORKDIR /
COPY --from=0  /client/bin/client_function /client_function
COPY ./tests/certs /certs
RUN apk update && apk add --no-cache iputils curl tcpdump busybox-extras

ENTRYPOINT ["./client_function"]