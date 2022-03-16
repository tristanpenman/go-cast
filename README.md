# GoCast

This project is intended to offer a complete implementation of the [Google Cast](https://en.wikipedia.org/wiki/Google_Cast) protocol, written in Go, with custom Sender and Receiver apps using [FFmpeg](https://ffmpeg.org/) as a frontend.

## Design

GoCast has been designed to be quite simple, while following many of the concepts that inherent to the Google Cast protocol. Namely, we preserve the concept of a Device that exists independently of the receiver apps that actually handle messages and display mirrored content.

The following architecture diagram shows an overview of how these concepts relate to one another:

![Architecture Diagram](./doc/architecture.drawio.png)

In terms of actual code, these concepts are implemented as Go structs.

## Usage

### Discovery App

The `discovery` app allows you to locate Google Cast devices on your network, using mDNS.

To run the discovery app in your local dev environment:

    go run cmd/discovery/*.go

Or to build an executable in `./bin/discovery`:

    go build -o ./bin/discovery ./cmd/discovery

### Receiver App

The `receiver` app will start a Google Cast receiver, which can then be cast to from compatible senders.  

To run the receiver in your local dev environment:

    go run cmd/receiver/*.go --cert-manifest=<path>

Or to build an executable in `./bin/receiver`:

    go build -o ./bin/receiver ./cmd/receiver

The `receiver` app can also be run in relay mode, where it will handle device authentication, but otherwise forward all messages and data to another host.

### Sender App

To run the sender in your local dev environment:

    go run cmd/sender/*.go

Or to build an executable in `./bin/sender`:

    go build -o ./bin/sender ./cmd/sender

### Cert Manifest

Before running the Receiver app, you will need to create (or otherwise obtain) a valid _certificate manifest_ file. A cert manifest is a JSON document containing the certificate and private key to be used TLS connections, and additional information used for Chromecast device authentication.

An example manifest is included in [etc/cert-manifest.json](./etc/cert-manifest.json). Note: This file does not include the fields required for device authentication.

## Protobuf

The Chromecast protocol relies on message types defined in protobuf format. The cast_channel.proto file in internal/message has been borrowed from the Chromium source code. To regenerate the Go bindings, run the following command:   

    protoc --go_opt=paths=source_relative --go_out=. ./internal/proto/cast_channel.proto
