package descriptor

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/jhump/protoreflect/desc"
	options "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func HandleFile(file *desc.FileDescriptor, fn func(meth *MethodWithBindings) error) error {
	for _, svc := range file.GetServices() {
		for _, md := range svc.GetMethods() {
			var optsList []*options.HttpRule
			opts, err := extractAPIOptions(md.AsMethodDescriptorProto())
			if err != nil {
				glog.Errorf("Failed to extract HttpRule from %s.%s: %v", svc.GetName(), md.GetName(), err)
				return err
			}
			// 暂时先不管
			// optsList := r.LookupExternalHTTPRules((&Method{Service: svc, MethodDescriptorProto: md}).FQMN())
			if opts != nil {
				optsList = append(optsList, opts)
			}
			if len(optsList) == 0 {
				continue
				// if r.generateUnboundMethods {
				// 	defaultOpts, err := defaultAPIOptions(svc, md)
				// 	if err != nil {
				// 		glog.Errorf("Failed to generate default HttpRule from %s.%s: %v", svc.GetName(), md.GetName(), err)
				// 		return err
				// 	}
				// 	optsList = append(optsList, defaultOpts)
			} else {
				// logFn := glog.V(1).Infof
				// if r.warnOnUnboundMethods {
				// 	logFn = glog.Warningf
				// }
				// logFn("No HttpRule found for method: %s.%s", svc.GetName(), md.GetName())
			}

			meth, err := newMethod(md, optsList)
			if err != nil {
				return err
			}
			if err := fn(meth); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractAPIOptions(meth *descriptorpb.MethodDescriptorProto) (*options.HttpRule, error) {
	if meth.Options == nil {
		return nil, nil
	}
	if !proto.HasExtension(meth.Options, options.E_Http) {
		return nil, nil
	}
	ext := proto.GetExtension(meth.Options, options.E_Http)
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return nil, fmt.Errorf("extension is %T; want an HttpRule", ext)
	}
	return opts, nil
}

// func defaultAPIOptions(svc *desc.ServiceDescriptor, md *descriptorpb.MethodDescriptorProto) (*options.HttpRule, error) {
// 	// FQSN prefixes the service's full name with a '.', e.g.: '.example.ExampleService'
// 	fqsn := strings.TrimPrefix(svc.AsServiceDescriptorProto().FQSN(), ".")

// 	// This generates an HttpRule that matches the gRPC mapping to HTTP/2 described in
// 	// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
// 	// i.e.:
// 	//   * method is POST
// 	//   * path is "/<service name>/<method name>"
// 	//   * body should contain the serialized request message
// 	rule := &options.HttpRule{
// 		Pattern: &options.HttpRule_Post{
// 			Post: fmt.Sprintf("/%s/%s", fqsn, md.GetName()),
// 		},
// 		Body: "*",
// 	}
// 	return rule, nil
// }
