package descriptor

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/micro-easy/go-zero/tools/goctl/gateway/casing"
	"google.golang.org/protobuf/types/descriptorpb"
)

type FieldPathComponent struct {
	// Name is a name of the proto field which this component corresponds to.
	// TODO(yugui) is this necessary?
	Name string
	// Target is the proto field which this component corresponds to.
	Target *desc.FieldDescriptor
}

// AssignableExpr returns an assignable expression in go for this field.
func (c FieldPathComponent) AssignableExpr() string {
	return casing.Camel(c.Name)
}

// ValueExpr returns an expression in go for this field.
func (c FieldPathComponent) ValueExpr() string {
	if proto2(c.Target.GetFile()) {
		return fmt.Sprintf("Get%s()", casing.Camel(c.Name))
	}
	return casing.Camel(c.Name)
}

// FieldPath is a path to a field from a request message.
type FieldPath []FieldPathComponent

// String returns a string representation of the field path.
func (p FieldPath) String() string {
	var components []string
	for _, c := range p {
		components = append(components, c.Name)
	}
	return strings.Join(components, ".")
}

// IsNestedProto3 indicates whether the FieldPath is a nested Proto3 path.
func (p FieldPath) IsNestedProto3() bool {
	// !p[0].Target.Message.File.proto2()
	if len(p) > 1 && !proto2(p[0].Target.GetFile()) {
		return true
	}
	return false
}

// IsOptionalProto3 indicates whether the FieldPath is a proto3 optional field.
func (p FieldPath) IsOptionalProto3() bool {
	if len(p) == 0 {
		return false
	}
	return p[0].Target.IsProto3Optional()
}

// AssignableExpr is an assignable expression in Go to be used to assign a value to the target field.
// It starts with "msgExpr", which is the go expression of the method request object.
func (p FieldPath) AssignableExpr(msgExpr string) string {
	l := len(p)
	if l == 0 {
		return msgExpr
	}

	var preparations []string
	components := msgExpr
	for i, c := range p {
		// We need to check if the target is not proto3_optional first.
		// Under the hood, proto3_optional uses oneof to signal to old proto3 clients
		// that presence is tracked for this field. This oneof is known as a "synthetic" oneof.
		targetfd := c.Target.AsFieldDescriptorProto()
		if !c.Target.IsProto3Optional() && targetfd.OneofIndex != nil {
			index := targetfd.OneofIndex
			msg := c.Target.GetMessageType()

			oneOfName := casing.Camel(msg.GetOneOfs()[*index].GetName())
			oneofFieldName := msg.GetName() + "_" + c.AssignableExpr()

			// if targetfd.ForcePrefixedName {
			// 	oneofFieldName = msg.File.Pkg() + "." + oneofFieldName
			// }

			components = components + "." + oneOfName
			s := `if %s == nil {
				%s =&%s{}
			} else if _, ok := %s.(*%s); !ok {
				return nil, metadata, status.Errorf(codes.InvalidArgument, "expect type: *%s, but: %%t\n",%s)
			}`

			preparations = append(preparations, fmt.Sprintf(s, components, components, oneofFieldName, components, oneofFieldName, oneofFieldName, components))
			components = components + ".(*" + oneofFieldName + ")"
		}

		if i == l-1 {
			components = components + "." + c.AssignableExpr()
			continue
		}
		components = components + "." + c.ValueExpr()
	}

	preparations = append(preparations, components)
	return strings.Join(preparations, "\n")
}

// Parameter is a parameter provided in http requests
type Parameter struct {
	// FieldPath is a path to a proto field which this parameter is mapped to.
	FieldPath
	// Target is the proto field which this parameter is mapped to.
	Target *desc.FieldDescriptor
	// Method is the method which this parameter is used for.
	Method *desc.MethodDescriptor
}

// ConvertFuncExpr returns a go expression of a converter function.
// The converter function converts a string into a value for the parameter.
func (p Parameter) ConvertFuncExpr() (string, error) {
	tbl := proto3ConvertFuncs
	if !p.IsProto2() && p.IsRepeated() {
		tbl = proto3RepeatedConvertFuncs
	} else if !p.IsProto2() && p.IsOptionalProto3() {
		tbl = proto3OptionalConvertFuncs
	} else if p.IsProto2() && !p.IsRepeated() {
		tbl = proto2ConvertFuncs
	} else if p.IsProto2() && p.IsRepeated() {
		tbl = proto2RepeatedConvertFuncs
	}
	typ := p.Target.GetType()
	conv, ok := tbl[typ]
	if !ok {
		targetFd := p.Target.AsFieldDescriptorProto()
		conv, ok = wellKnownTypeConv[targetFd.GetTypeName()]
	}
	if !ok {
		return "", fmt.Errorf("unsupported field type %s of parameter %s in %s.%s", typ, p.FieldPath, p.Method.GetService().GetName(), p.Method.GetName())
	}
	return conv, nil
}

// IsEnum returns true if the field is an enum type, otherwise false is returned.
func (p Parameter) IsEnum() bool {
	return p.Target.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM
}

// IsRepeated returns true if the field is repeated, otherwise false is returned.
func (p Parameter) IsRepeated() bool {
	return p.Target.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

// IsProto2 returns true if the field is proto2, otherwise false is returned.
func (p Parameter) IsProto2() bool {
	return proto2(p.Target.GetFile())
}

// Body describes a http (request|response) body to be sent to the (method|client).
// This is used in body and response_body options in google.api.HttpRule
type Body struct {
	// FieldPath is a path to a proto field which the (request|response) body is mapped to.
	// The (request|response) body is mapped to the (request|response) type itself if FieldPath is empty.
	FieldPath FieldPath
}

// AssignableExpr returns an assignable expression in Go to be used to initialize method request object.
// It starts with "msgExpr", which is the go expression of the method request object.
func (b Body) AssignableExpr(msgExpr string) string {
	return b.FieldPath.AssignableExpr(msgExpr)
}

// resolveFieldPath resolves "path" into a list of fieldDescriptor, starting from "msg".
func resolveFieldPath(msg *desc.MessageDescriptor, path string, isPathParam bool) ([]FieldPathComponent, error) {
	if path == "" {
		return nil, nil
	}

	root := msg
	var result []FieldPathComponent
	for i, c := range strings.Split(path, ".") {
		if i > 0 {
			// 前一个field是父级field
			f := result[i-1].Target
			switch f.GetType() {
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, descriptorpb.FieldDescriptorProto_TYPE_GROUP:
				msg = f.GetMessageType()
				// r.LookupMsg(msg.FQMN(), f.GetTypeName())
			default:
				return nil, fmt.Errorf("not an aggregate type: %s in %s", f.GetName(), path)
			}
		}

		f := msg.FindFieldByName(c)
		// f := lookupField(msg, c)
		if f == nil {
			return nil, fmt.Errorf("no field %q found in %s", path, root.GetName())
		}
		// if !(isPathParam || r.allowRepeatedFieldsInBody) && f.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		// 	return nil, fmt.Errorf("repeated field not allowed in field path: %s in %s", f.GetName(), path)
		// }
		if isPathParam && f.IsProto3Optional() {
			return nil, fmt.Errorf("optional field not allowed in field path: %s in %s", f.GetName(), path)
		}
		result = append(result, FieldPathComponent{Name: c, Target: f})
	}
	return result, nil
}

// IsWellKnownType returns true if the provided fully qualified type name is considered 'well-known'.
func IsWellKnownType(typeName string) bool {
	_, ok := wellKnownTypeConv[typeName]
	return ok
}

// proto2 determines if the syntax of the file is proto2.
func proto2(fd *desc.FileDescriptor) bool {
	file := fd.AsFileDescriptorProto()
	return file.Syntax == nil || file.GetSyntax() == "proto2"
}

var (
	proto3ConvertFuncs = map[descriptorpb.FieldDescriptorProto_Type]string{
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:  "runtime.Float64",
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:   "runtime.Float32",
		descriptorpb.FieldDescriptorProto_TYPE_INT64:   "runtime.Int64",
		descriptorpb.FieldDescriptorProto_TYPE_UINT64:  "runtime.Uint64",
		descriptorpb.FieldDescriptorProto_TYPE_INT32:   "runtime.Int32",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64: "runtime.Uint64",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32: "runtime.Uint32",
		descriptorpb.FieldDescriptorProto_TYPE_BOOL:    "runtime.Bool",
		descriptorpb.FieldDescriptorProto_TYPE_STRING:  "runtime.String",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		descriptorpb.FieldDescriptorProto_TYPE_BYTES:    "runtime.Bytes",
		descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "runtime.Uint32",
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:     "runtime.Enum",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "runtime.Int32",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "runtime.Int64",
		descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "runtime.Int32",
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "runtime.Int64",
	}

	proto3OptionalConvertFuncs = func() map[descriptorpb.FieldDescriptorProto_Type]string {
		result := make(map[descriptorpb.FieldDescriptorProto_Type]string)
		for typ, converter := range proto3ConvertFuncs {
			// TODO: this will use convert functions from proto2.
			//       The converters returning pointers should be moved
			//       to a more generic file.
			result[typ] = converter + "P"
		}
		return result
	}()

	// TODO: replace it with a IIFE
	proto3RepeatedConvertFuncs = map[descriptorpb.FieldDescriptorProto_Type]string{
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:  "runtime.Float64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:   "runtime.Float32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_INT64:   "runtime.Int64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_UINT64:  "runtime.Uint64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_INT32:   "runtime.Int32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64: "runtime.Uint64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32: "runtime.Uint32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_BOOL:    "runtime.BoolSlice",
		descriptorpb.FieldDescriptorProto_TYPE_STRING:  "runtime.StringSlice",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		descriptorpb.FieldDescriptorProto_TYPE_BYTES:    "runtime.BytesSlice",
		descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "runtime.Uint32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:     "runtime.EnumSlice",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "runtime.Int32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "runtime.Int64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "runtime.Int32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "runtime.Int64Slice",
	}

	proto2ConvertFuncs = map[descriptorpb.FieldDescriptorProto_Type]string{
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:  "runtime.Float64P",
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:   "runtime.Float32P",
		descriptorpb.FieldDescriptorProto_TYPE_INT64:   "runtime.Int64P",
		descriptorpb.FieldDescriptorProto_TYPE_UINT64:  "runtime.Uint64P",
		descriptorpb.FieldDescriptorProto_TYPE_INT32:   "runtime.Int32P",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64: "runtime.Uint64P",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32: "runtime.Uint32P",
		descriptorpb.FieldDescriptorProto_TYPE_BOOL:    "runtime.BoolP",
		descriptorpb.FieldDescriptorProto_TYPE_STRING:  "runtime.StringP",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		// FieldDescriptorProto_TYPE_BYTES
		// TODO(yugui) Handle bytes
		descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "runtime.Uint32P",
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:     "runtime.EnumP",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "runtime.Int32P",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "runtime.Int64P",
		descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "runtime.Int32P",
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "runtime.Int64P",
	}

	proto2RepeatedConvertFuncs = map[descriptorpb.FieldDescriptorProto_Type]string{
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:  "runtime.Float64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT:   "runtime.Float32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_INT64:   "runtime.Int64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_UINT64:  "runtime.Uint64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_INT32:   "runtime.Int32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64: "runtime.Uint64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32: "runtime.Uint32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_BOOL:    "runtime.BoolSlice",
		descriptorpb.FieldDescriptorProto_TYPE_STRING:  "runtime.StringSlice",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		// FieldDescriptorProto_TYPE_BYTES
		// TODO(maros7) Handle bytes
		descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "runtime.Uint32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:     "runtime.EnumSlice",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "runtime.Int32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "runtime.Int64Slice",
		descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "runtime.Int32Slice",
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "runtime.Int64Slice",
	}

	wellKnownTypeConv = map[string]string{
		".google.protobuf.Timestamp":   "runtime.Timestamp",
		".google.protobuf.Duration":    "runtime.Duration",
		".google.protobuf.StringValue": "runtime.StringValue",
		".google.protobuf.FloatValue":  "runtime.FloatValue",
		".google.protobuf.DoubleValue": "runtime.DoubleValue",
		".google.protobuf.BoolValue":   "runtime.BoolValue",
		".google.protobuf.BytesValue":  "runtime.BytesValue",
		".google.protobuf.Int32Value":  "runtime.Int32Value",
		".google.protobuf.UInt32Value": "runtime.UInt32Value",
		".google.protobuf.Int64Value":  "runtime.Int64Value",
		".google.protobuf.UInt64Value": "runtime.UInt64Value",
	}
)
