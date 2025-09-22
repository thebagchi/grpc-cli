# gRPC CLI Tool

A command-line tool for interacting with gRPC services using server reflection or proto files.

## Features

- **Server Reflection Support**: Automatically discover services without proto files
- **Proto File Support**: Use local proto files with custom import paths  
- **List Methods**: Display all available gRPC methods
- **Make gRPC Calls**: Execute RPC calls with JSON payloads

## Building

```bash
go build -o grpc_cli.bin
```

## Usage

```bash
./grpc_cli.bin [options]
```

### Command Line Options

| Option | Type | Description |
|--------|------|-------------|
| `-h` | string | gRPC server host address (e.g: localhost:12345). Default: *(localhost:12345)* |
| `-p` | string | Input proto file for the gRPC server, reflection will be used if omitted. Default: *(empty)* |
| `-i` | value | Include path(s) for compiling proto (can be specified multiple times). Default: *(none)* |
| `-l` | bool | List available methods. Default: *(false)* |
| `-m` | string | Method name for RPC in format [package].[service].[rpc]. Default: *(empty)* |
| `-d` | string | JSON data used as input for RPC. Default: *(empty)* |

### Examples

#### List all available methods using reflection
```bash
./grpc_cli.bin -h "localhost:9090" -l
```

#### List methods using a proto file
```bash
./grpc_cli.bin -p "./proto/service.proto" -i "./proto" -l
```

#### Make an RPC call using reflection
```bash
./grpc_cli.bin -h "localhost:9090" \
  -m "mypackage.MyService.GetUser" \
  -d '{"id": "123", "name": "John"}'
```

#### Make an RPC call using proto file with custom imports
```bash
./grpc_cli.bin -h "localhost:9090" \
  -p "./proto/service.proto" \
  -i "./proto" \
  -i "./proto/common" \
  -m "mypackage.MyService.GetUser" \
  -d '{"id": "123", "name": "John"}'
```

#### Connect to remote server
```bash
./grpc_cli.bin -h "grpc.example.com:443" -l
```

#### View help
```bash
./grpc_cli.bin -help
```

## How it works

### Mode 1: Server Reflection (Default)
1. **Connects to gRPC server** using the specified host address
2. **Uses server reflection** to discover available services
3. **Lists available services** and retrieves file descriptors
4. **Displays service information** or executes RPC calls

### Mode 2: Proto File Compilation
1. **Compiles proto files** using specified include paths
2. **Loads service definitions** from compiled descriptors
3. **Lists methods** or executes RPC calls based on proto definitions

### Method Name Format
Methods must be specified in the format: `[package].[service].[rpc]`

Examples:
- `grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo`
- `myapi.UserService.GetUser`
- `com.example.OrderService.CreateOrder`

### JSON Input Format
The `-d` parameter accepts JSON data that matches the input message structure of the target RPC method.

Example:
```json
{
  "id": "user123",
  "name": "John Doe",
  "email": "john@example.com"
}
```

## Requirements

- Go 1.24.4 or later
- For reflection mode: Target gRPC server must have reflection enabled
- For proto mode: Proto files and any dependencies must be accessible
- Network connectivity to the target gRPC server

## Dependencies

- `google.golang.org/grpc` - gRPC Go library
- `google.golang.org/protobuf` - Protocol Buffers Go library
- `github.com/jhump/protoreflect` - Protocol buffer reflection utilities

## Output

The tool produces detailed output including:
- List of available services and methods (with `-l` flag)
- Service method details and signatures
- RPC call results in JSON format
- Error messages with helpful context

All JSON output is formatted with indentation for readability.