//go:build tools
// +build tools

package main

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)

// Check for upgrades and do upgrade ...
//go:generate go list -u -m all
//go get -u all
