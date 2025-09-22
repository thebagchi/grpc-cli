package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/protoutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ListString struct {
	items []string
}

func (m *ListString) String() string {
	return strings.Join(m.items, ",")
}

func (m *ListString) Set(v string) error {
	m.items = append(m.items, v)
	return nil
}

func MakeLS() *ListString {
	return &ListString{
		items: make([]string, 0),
	}
}

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

func FindServiceDescriptor(desc protoreflect.FileDescriptor, name string) (protoreflect.ServiceDescriptor, error) {
	if strings.HasPrefix(name, fmt.Sprintf("%s.", desc.Package())) {
		name = strings.TrimPrefix(name, fmt.Sprintf("%s.", desc.Package()))
		services := desc.Services()
		if sdesc := services.ByName(protoreflect.Name(name)); sdesc != nil {
			return sdesc, nil
		}
	}
	return nil, fmt.Errorf("%s has no .%s field", desc.FullName(), name)
}

func MakeCall(conn *grpc.ClientConn, descriptors *descriptorpb.FileDescriptorSet, svc, rpc string, data string) (string, error) {
	files, err := protodesc.NewFiles(descriptors)
	if nil != err {
		return "", nil
	}
	var (
		service   protoreflect.ServiceDescriptor = nil
		procedure protoreflect.MethodDescriptor  = nil
		found     bool                           = false
	)
	files.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
		tmp, err := FindServiceDescriptor(descriptor, svc)
		if nil == err {
			service = tmp
			found = true
		}
		return true
	})
	if !found {
		return "", fmt.Errorf("service not found method: %s service: %s", rpc, svc)
	}
	procedure = service.Methods().ByName(protoreflect.Name(rpc))
	if nil == procedure {
		return "", fmt.Errorf("service not found method: %s service: %s", rpc, svc)
	}
	if nil == procedure.Input() || nil == procedure.Output() {
		return "", fmt.Errorf("invalid service method: %s service: %s", rpc, svc)
	}
	var (
		request  = dynamicpb.NewMessage(procedure.Input())
		response = dynamicpb.NewMessage(procedure.Output())
	)
	err = protojson.Unmarshal([]byte(data), request)
	if nil != err {
		return "", err
	}
	err = conn.Invoke(context.Background(), fmt.Sprintf("/%s/%s", svc, rpc), request, response)
	if nil != err {
		return "", err
	}
	buffer, err := protojson.Marshal(response)
	if nil != err {
		return "", err
	}
	return string(buffer), nil
}

func makeFD(file linker.File) []*descriptorpb.FileDescriptorProto {
	var (
		descs = make([]*descriptorpb.FileDescriptorProto, 0)
		desc  = protoutil.ProtoFromFileDescriptor(file)
	)
	for _, depends := range desc.Dependency {
		temp := makeFD(file.FindImportByPath(depends))
		descs = append(descs, temp...)
	}
	descs = append(descs, desc)
	return descs
}

func compileProto(path string, imports []string) *descriptorpb.FileDescriptorSet {
	compiler := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: imports,
		}),
	}
	fdescs := new(descriptorpb.FileDescriptorSet)
	files, err := compiler.Compile(context.Background(), path)
	if nil != err {
		fmt.Println("Error: ", err)
		return nil
	}
	fdescs.File = makeFD(files.FindFileByPath(path))
	if len(fdescs.File) > 0 {
		return fdescs
	}
	return nil
}

func loadReflection(conn *grpc.ClientConn, host string) *descriptorpb.FileDescriptorSet {
	client := reflectpb.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(context.Background())
	if nil != err {
		fmt.Println("Error: ", err)
		return nil
	}
	services := make([]string, 0)
	for {
		request := &reflectpb.ServerReflectionRequest{
			Host:           host,
			MessageRequest: new(reflectpb.ServerReflectionRequest_ListServices),
		}
		err := stream.Send(request)
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		response, err := stream.Recv()
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		svcs := response.GetListServicesResponse()
		for _, service := range svcs.GetService() {
			services = append(services, service.Name)
		}
		break
	}
	if len(services) > 0 {
		fdescs := new(descriptorpb.FileDescriptorSet)
		for _, service := range services {
			request := &reflectpb.ServerReflectionRequest{
				Host: host,
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
			if nil != err {
				fmt.Println("Error: ", err)
			}
			decs := response.GetFileDescriptorResponse()
			for _, buffer := range decs.GetFileDescriptorProto() {
				descriptor := new(descriptorpb.FileDescriptorProto)
				proto.Unmarshal(buffer, descriptor)
				AddDescriptorToSet(fdescs, descriptor)
			}
		}
		if len(fdescs.File) > 0 {
			return fdescs
		}
		return fdescs
	}
	return nil
}

func listMethods(fdescs *descriptorpb.FileDescriptorSet) {
	methods := make([]string, 0)
	for {
		files, err := protodesc.NewFiles(fdescs)
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		files.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
			services := descriptor.Services()
			for i := 0; i < services.Len(); i++ {
				service := services.Get(i)
				if !strings.HasPrefix(string(service.FullName()), "grpc.reflection") {
					for j := 0; j < service.Methods().Len(); j++ {
						methods = append(methods, string(service.Methods().Get(j).FullName()))
					}
				}
			}
			return true
		})
		break
	}
	for _, method := range methods {
		fmt.Println(" = ", method)
	}
}

func main() {
	var (
		host                                          = "localhost:12345"
		protobuf                                      = ""
		methods                                       = false
		method                                        = ""
		data                                          = ""
		imports                                       = MakeLS()
		fdescs        *descriptorpb.FileDescriptorSet = nil
		useReflection                                 = false
	)

	flag.StringVar(&host, "h", "localhost:12345", "grpc server host address (e.g: localhost:12345)")
	flag.StringVar(&protobuf, "p", "", "input proto file for the grpc server, reflection will be used if omitted")
	flag.BoolVar(&methods, "l", false, "list methods")
	flag.StringVar(&method, "m", "", "method name for rpc [package].[service].[rpc]")
	flag.StringVar(&data, "d", "", "json data used as input for rpc")
	flag.Var(imports, "i", "include path(s) for compiling proto")
	flag.Parse()

	if len(protobuf) != 0 {
		fdescs = compileProto(protobuf, imports.items)
	}
	if nil == fdescs {
		useReflection = true
	}
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if nil != err {
		fmt.Println("Error: ", err)
		return
	}
	if useReflection {
		fdescs = loadReflection(conn, host)
	}
	if nil == fdescs {
		fmt.Println("Error: cannot continue unable to load file descriptors")
		return
	}
	if methods {
		listMethods(fdescs)
		return
	}
	for {
		items := strings.Split(method, ".")
		if len(items) < 3 {
			fmt.Println("Error: invalid method name want [package].[service].[rpc] (e.g: grpc.reflection.v1.ServerReflection.ServerReflectionInfo)")
			break
		}
		var (
			service   = strings.Join(items[:len(items)-1], ".")
			procedure = items[len(items)-1]
		)
		data, err := MakeCall(conn, fdescs, service, procedure, data)
		if nil != err {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println(data)
		break
	}
}
