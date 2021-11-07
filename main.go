package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func AddDescriptorToSet(descriptors *descriptorpb.FileDescriptorSet, descriptor *descriptorpb.FileDescriptorProto) {
	if descriptor.GetName() == "src/proto/grpc/reflection/v1alpha/reflection.proto" {
		return
	}
	found := false
	for _, file := range descriptors.File {
		if file.GetName() == descriptor.GetName() {
			found = true
			break
		}
	}
	if !found {
		descriptors.File = append(descriptors.File, descriptor)
	}
}

func FindFieldDescriptor(desc protoreflect.MessageDescriptor, name string) (protoreflect.FieldDescriptor, error) {
	if fdesc := desc.Fields().ByName(protoreflect.Name(name)); fdesc != nil {
		return fdesc, nil
	}
	return nil, fmt.Errorf("%s has no .%s field", desc.FullName(), name)
}

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

	encoder := protojson.MarshalOptions{
		Indent: "    ",
	}
	fmt.Println("=====================================================================================================")
	// List Services ...
	services := make([]string, 0)
	for {
		request := &reflectpb.ServerReflectionRequest{
			Host:           "localhost:12345",
			MessageRequest: new(reflectpb.ServerReflectionRequest_ListServices),
		}
		err := stream.Send(request)
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		response, err := stream.Recv()
		fmt.Println(encoder.Format(response))
		svcs := response.GetListServicesResponse()
		for _, service := range svcs.GetService() {
			services = append(services, service.Name)
		}
		break
	}
	fmt.Println("=====================================================================================================")
	descriptors := new(descriptorpb.FileDescriptorSet)
	// Prepare FileDescriptorSet ...
	if len(services) > 0 {
		for _, service := range services {
			request := &reflectpb.ServerReflectionRequest{
				Host: "localhost:12345",
				MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{
					FileContainingSymbol: service,
				},
			}
			err := stream.Send(request)
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			response, err := stream.Recv()
			fmt.Println(encoder.Format(response))
			decs := response.GetFileDescriptorResponse()
			for _, buffer := range decs.GetFileDescriptorProto() {
				descriptor := new(descriptorpb.FileDescriptorProto)
				proto.Unmarshal(buffer, descriptor)
				fmt.Println(descriptor.GetName())
				AddDescriptorToSet(descriptors, descriptor)
			}
		}
	}
	fmt.Println("=====================================================================================================")
	fmt.Println(encoder.Format(descriptors))
	files := make([]string, 0)
	// Prepare List Of Files ...
	for _, descriptor := range descriptors.GetFile() {
		files = append(files, descriptor.GetName())
	}
	if len(files) > 0 {
		for _, file := range files {
			request := &reflectpb.ServerReflectionRequest{
				Host: "localhost:12345",
				MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{
					FileByFilename: file,
				},
			}
			err := stream.Send(request)
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			response, err := stream.Recv()
			fmt.Println(encoder.Format(response))
		}
	}
	fmt.Println("=====================================================================================================")
	// List Messages ...
	messages := make([]protoreflect.MessageDescriptor, 0)
	for {
		files, err := protodesc.NewFiles(descriptors)
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		files.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
			for i := 0; i < descriptor.Messages().Len(); i++ {
				descriptor := descriptor.Messages().Get(i)
				messages = append(messages, descriptor)
			}
			return true
		})
		break
	}
	fmt.Println("=====================================================================================================")
	// Populate Message
	// Refer to https://github.com/thebagchi/grpc_async/blob/master/proto/rpc.proto
	for {
		var descriptor protoreflect.MessageDescriptor = nil
		for _, msg := range messages {
			if msg.FullName() == "rpc.SampleMessage" {
				descriptor = msg
				break
			}
		}
		message := dynamicpb.NewMessage(descriptor)
		// Set Members
		{
			fd, err := FindFieldDescriptor(descriptor, "string_value")
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			message.Set(fd, protoreflect.ValueOfString("Hello World!!!"))
		}
		{
			fd, err := FindFieldDescriptor(descriptor, "integer_value")
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			message.Set(fd, protoreflect.ValueOfInt64(54321))
		}
		{
			fd, err := FindFieldDescriptor(descriptor, "boolean_value")
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			message.Set(fd, protoreflect.ValueOfBool(false))
		}
		{
			buffer, err := proto.Marshal(message)
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			fmt.Println(hex.Dump(buffer))
		}
		{
			buffer, err := protojson.Marshal(message)
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			fmt.Println(string(buffer))
		}
		{
			json := `
{"stringValue":"Hello World!!!","integerValue":"54321","booleanValue":false}
`
			message := dynamicpb.NewMessage(descriptor)
			err := protojson.Unmarshal([]byte(json), message)
			if nil != err {
				fmt.Println("Error: ", err)
				break
			}
			fmt.Println(message)
		}
		break
	}
	fmt.Println("=====================================================================================================")
	// Make GRPC Call
	fmt.Println("=====================================================================================================")
}
