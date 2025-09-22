//go:build tools
// +build tools

package main

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)

// Run this for checking if there are upgrades ...
//go:generate go list -u -m all

// Run this for upgrading all dependencies ...
//go get -u all
