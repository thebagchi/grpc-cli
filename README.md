# gRPC CLI Tool

A command-line tool for interacting with gRPC services using server reflection.

## Features

- List available gRPC services
- Retrieve service descriptors and file descriptors
- Make gRPC calls with JSON payloads
- Support for dynamic message creation and manipulation
- Server reflection support

## Building

```bash
go build -o grpc-cli
```

## Usage

```bash
./grpc-cli [options]
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-host string` | gRPC server host address (e.g., localhost:12345) | `localhost:12345` |

### Examples

#### Using default host (localhost:12345)
```bash
./grpc-cli
```

#### Connecting to a custom host
```bash
./grpc-cli -host "myserver:9090"
./grpc-cli -host "192.168.1.100:8080"
./grpc-cli -host "grpc.example.com:443"
```

#### View help
```bash
./grpc-cli -help
```

## What the tool does

1. **Connects to gRPC server** using the specified host address
2. **Lists available services** using server reflection
3. **Retrieves file descriptors** for all discovered services
4. **Displays service information** in JSON format
5. **Demonstrates message creation** and manipulation
6. **Makes sample gRPC calls** (configured for `rpc.SampleSvc.RPC_1`)

## Requirements

- Go 1.17 or later
- Target gRPC server must have reflection enabled
- Network connectivity to the target gRPC server

## Dependencies

- `google.golang.org/grpc` - gRPC Go library
- `google.golang.org/protobuf` - Protocol Buffers Go library
- `grpc-ecosystem/grpc-gateway/v2` - gRPC Gateway

## Output

The tool produces detailed output including:
- List of available services
- File descriptor information
- Service method details
- Sample message creation and serialization
- gRPC call results

All output is formatted with indented JSON for readability.