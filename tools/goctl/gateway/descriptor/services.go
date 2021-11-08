package descriptor

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/jhump/protoreflect/desc"
	"github.com/micro-easy/go-zero/tools/goctl/gateway/casing"
	"github.com/micro-easy/go-zero/tools/goctl/gateway/httprule"
	options "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/types/descriptorpb"
)

func newParam(meth *desc.MethodDescriptor, path string) (Parameter, error) {
	msg := meth.GetInputType()
	fields, err := resolveFieldPath(msg, path, true)
	if err != nil {
		return Parameter{}, err
	}
	l := len(fields)
	if l == 0 {
		return Parameter{}, fmt.Errorf("invalid field access list for %s", path)
	}
	// 从最后一个开始遍历
	target := fields[l-1].Target
	switch target.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		// TypeName 没有暴露出来，后面再看
		targetFd := target.AsFieldDescriptorProto()
		if IsWellKnownType(targetFd.GetTypeName()) {
			glog.V(2).Infoln("found well known aggregate type:", target)
		} else {
			return Parameter{}, fmt.Errorf("%s.%s: %s is a protobuf message type. Protobuf message types cannot be used as path parameters, use a scalar value type (such as string) instead", meth.GetService().GetName(), meth.GetName(), path)
		}
	}
	return Parameter{
		FieldPath: FieldPath(fields),
		Method:    meth,
		Target:    fields[l-1].Target,
	}, nil
}

func newBody(meth *desc.MethodDescriptor, path string) (*Body, error) {
	msg := meth.GetInputType()
	switch path {
	case "":
		return nil, nil
	case "*":
		return &Body{FieldPath: nil}, nil
	}
	fields, err := resolveFieldPath(msg, path, false)
	if err != nil {
		return nil, err
	}
	return &Body{FieldPath: FieldPath(fields)}, nil
}

func newResponse(meth *desc.MethodDescriptor, path string) (*Body, error) {
	msg := meth.GetOutputType()
	switch path {
	case "", "*":
		return nil, nil
	}
	fields, err := resolveFieldPath(msg, path, false)
	if err != nil {
		return nil, err
	}
	return &Body{FieldPath: FieldPath(fields)}, nil
}

// Binding describes how an HTTP endpoint is bound to a gRPC method.
type Binding struct {
	// Method is the method which the endpoint is bound to.
	Method *desc.MethodDescriptor
	// Index is a zero-origin index of the binding in the target method
	Index int
	// PathTmpl is path template where this method is mapped to.
	PathTmpl httprule.Template
	// HTTPMethod is the HTTP method which this method is mapped to.
	HTTPMethod string
	// PathParams is the list of parameters provided in HTTP request paths.
	PathParams []Parameter
	// Body describes parameters provided in HTTP request body.
	Body *Body
	// ResponseBody describes field in response struct to marshal in HTTP response body.
	ResponseBody *Body
}

// ExplicitParams returns a list of explicitly bound parameters of "b",
// i.e. a union of field path for body and field paths for path parameters.
func (b *Binding) ExplicitParams() []string {
	var result []string
	if b.Body != nil {
		result = append(result, b.Body.FieldPath.String())
	}
	for _, p := range b.PathParams {
		result = append(result, p.FieldPath.String())
	}
	return result
}

// hasEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to a enum proto field and that the enum proto field is or isn't repeated
// based on the provided 'repeated' parameter.
func (b *Binding) HasEnumPathParam() bool {
	for _, p := range b.PathParams {
		if p.IsEnum() && !p.IsRepeated() {
			return true
		}
	}
	return false
}

func (b *Binding) HasRepeatedEnumPathParam() bool {
	for _, p := range b.PathParams {
		if p.IsEnum() && p.IsRepeated() {
			return true
		}
	}
	return false
}

// LookupEnum looks up a enum type by path parameter.
func (b *Binding) LookupEnum(p Parameter) *desc.EnumDescriptor {
	// p.GetFile().
	return p.Target.GetEnumType()
}

// FieldMaskField returns the golang-style name of the variable for a FieldMask, if there is exactly one of that type in
// the message. Otherwise, it returns an empty string.
func (b *Binding) FieldMaskField() string {
	var fieldMaskField *desc.FieldDescriptor
	// var fieldMaskField *descriptor.Field
	for _, f := range b.Method.GetInputType().GetFields() {
		if f.AsFieldDescriptorProto().GetTypeName() == ".google.protobuf.FieldMask" {
			// if there is more than 1 FieldMask for this request, then return none
			if fieldMaskField != nil {
				return ""
			}
			fieldMaskField = f
		}
	}
	if fieldMaskField != nil {
		return casing.Camel(fieldMaskField.GetName())
	}
	return ""
}

func (b *Binding) GetRepeatedPathParamSeparator() rune {
	return ','
}

type MethodWithBindings struct {
	*desc.MethodDescriptor
	Bindings []*Binding
}

func newMethod(md *desc.MethodDescriptor, optsList []*options.HttpRule) (*MethodWithBindings, error) {
	meth := &MethodWithBindings{
		MethodDescriptor: md,
	}
	newBinding := func(opts *options.HttpRule, idx int) (*Binding, error) {
		var (
			httpMethod   string
			pathTemplate string
		)
		switch {
		case opts.GetGet() != "":
			httpMethod = "GET"
			pathTemplate = opts.GetGet()
			if opts.Body != "" {
				return nil, fmt.Errorf("must not set request body when http method is GET: %s", md.GetName())
			}

		case opts.GetPut() != "":
			httpMethod = "PUT"
			pathTemplate = opts.GetPut()

		case opts.GetPost() != "":
			httpMethod = "POST"
			pathTemplate = opts.GetPost()

		case opts.GetDelete() != "":
			httpMethod = "DELETE"
			pathTemplate = opts.GetDelete()
			// if opts.Body != "" && !r.allowDeleteBody {
			// 	return nil, fmt.Errorf("must not set request body when http method is DELETE except allow_delete_body option is true: %s", md.GetName())
			// }

		case opts.GetPatch() != "":
			httpMethod = "PATCH"
			pathTemplate = opts.GetPatch()

		case opts.GetCustom() != nil:
			custom := opts.GetCustom()
			httpMethod = custom.Kind
			pathTemplate = custom.Path

		default:
			glog.V(1).Infof("No pattern specified in google.api.HttpRule: %s", md.GetName())
			return nil, nil
		}

		parsed, err := httprule.Parse(pathTemplate)
		if err != nil {
			return nil, err
		}
		tmpl := parsed.Compile()

		if md.AsMethodDescriptorProto().GetClientStreaming() && len(tmpl.Fields) > 0 {
			return nil, fmt.Errorf("cannot use path parameter in client streaming")
		}

		b := &Binding{
			Method:     md,
			Index:      idx,
			PathTmpl:   tmpl,
			HTTPMethod: httpMethod,
		}

		for _, f := range tmpl.Fields {
			param, err := newParam(md, f)
			if err != nil {
				return nil, err
			}
			b.PathParams = append(b.PathParams, param)
		}

		b.Body, err = newBody(md, opts.Body)
		if err != nil {
			return nil, err
		}

		b.ResponseBody, err = newResponse(md, opts.ResponseBody)
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	applyOpts := func(opts *options.HttpRule) error {
		b, err := newBinding(opts, len(meth.Bindings))
		if err != nil {
			return err
		}

		if b != nil {
			meth.Bindings = append(meth.Bindings, b)
		}
		for _, additional := range opts.GetAdditionalBindings() {
			if len(additional.AdditionalBindings) > 0 {
				return fmt.Errorf("additional_binding in additional_binding not allowed: %s.%s", meth.GetService().GetName(), meth.GetName())
			}
			b, err := newBinding(additional, len(meth.Bindings))
			if err != nil {
				return err
			}
			meth.Bindings = append(meth.Bindings, b)
		}

		return nil
	}

	for _, opts := range optsList {
		if err := applyOpts(opts); err != nil {
			return nil, err
		}
	}

	return meth, nil
}
