package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	conn, err := grpc.Dial("localhost:12345", grpc.WithInsecure())
	if nil != err {
		fmt.Println("Error: ", err)
		return
	}
	client := reflectpb.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(context.Background())
	if nil != err {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("=====================================================================================================")
	encoder := protojson.MarshalOptions {
		Indent: "    ",
	}
	// List Services ...
	for {
		request := &reflectpb.ServerReflectionRequest{
			Host: "localhost:12345",
			MessageRequest: new(reflectpb.ServerReflectionRequest_ListServices),
		}
		err := stream.Send(request)
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		response, err:= stream.Recv()
		fmt.Println(encoder.Format(response))
		break
	}
	fmt.Println("=====================================================================================================")
}
