// Package proto provides protobuf compatibility helpers for Argo CD.
//
// The primary export is LegacyMessage, which bridges gogo/protobuf-generated types
// (protoiface.MessageV1) into the google.golang.org/protobuf (APIv2) world by
// satisfying the protoreflect.Message interface. This is the same role played
// internally by protoadapt.MessageV2Of / legacyWrapMessage inside the protobuf
// runtime, but implemented explicitly here to avoid the re-entrant mutex deadlock
// described below.
//
// # Why not use protoadapt.MessageV2Of?
//
// protoadapt.MessageV2Of delegates to internal/impl.legacyWrapMessage which calls
// aberrantLoadMessageDesc under a single process-wide non-reentrant mutex
// (aberrantMessageDescLock). The v1alpha1 types reference each other as nested
// fields (e.g. Application → ApplicationSpec → ApplicationSource …). When the
// descriptor for one type is being synthesised under that lock, encountering a
// nested field whose type already implements proto.Message (MessageV2) triggers a
// recursive call to ProtoReflect().Descriptor() on that nested type, which in turn
// attempts to re-acquire the same mutex — causing a deadlock on the very first
// proto.Marshal call.
//
// Using a bespoke legacyMessage that only implements the fast-path ProtoMethods
// (Marshal / Unmarshal / Size / CheckInitialized) sidesteps descriptor synthesis
// entirely. The gRPC binary codec and grpc-gateway HTTP path both use only this
// fast path; neither requires a compiled file descriptor.
package proto

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

// Serializable is the fast-path serialisation interface satisfied by every
// gogo/protobuf-generated type via its Marshal(), Unmarshal(), and Size() methods.
type Serializable interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	Size() int
}

// LegacyMessage returns a protoreflect.Message that provides correct proto binary
// serialisation for any gogo-generated type by delegating to its Marshal/Unmarshal/Size
// methods.
//
// self must be the proto.Message wrapper (typically the same pointer as ser, but may
// differ when a wrapper type is needed for compile-time interface satisfaction).
// It is returned unchanged by Interface() so that grpc-gateway can recover the
// original value after passing it around as proto.Message.
//
// ser is the value that provides the gogo-generated Marshal/Unmarshal/Size methods.
// It may be the same pointer as self or an embedded field of it.
//
// Only ProtoMethods(), Interface(), and IsValid() are fully implemented.
// All field-reflection methods (Range, Get, Descriptor, …) intentionally panic:
// they require a compiled proto file descriptor that is not available for gogo types,
// and they are never invoked by the gRPC binary codec or the grpc-gateway HTTP path.
func LegacyMessage(self protoreflect.ProtoMessage, ser Serializable) protoreflect.Message {
	// Detect typed-nil pointers: a nil *T passed as an interface is non-nil at
	// the interface level but should be considered invalid (IsValid → false).
	valid := self != nil
	if valid {
		if rv := reflect.ValueOf(self); rv.Kind() == reflect.Ptr {
			valid = !rv.IsNil()
		}
	}
	lm := &legacyMessage{self: self, valid: valid}
	lm.methods.Flags = protoiface.SupportMarshalDeterministic | protoiface.SupportUnmarshalDiscardUnknown
	lm.methods.Size = func(in protoiface.SizeInput) protoiface.SizeOutput {
		return protoiface.SizeOutput{Size: ser.Size()}
	}
	lm.methods.Marshal = func(in protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		b, err := ser.Marshal()
		if err != nil {
			return protoiface.MarshalOutput{}, err
		}
		return protoiface.MarshalOutput{Buf: append(in.Buf, b...)}, nil
	}
	lm.methods.Unmarshal = func(in protoiface.UnmarshalInput) (protoiface.UnmarshalOutput, error) {
		return protoiface.UnmarshalOutput{}, ser.Unmarshal(in.Buf)
	}
	// CheckInitialized prevents proto.Marshal from falling back to the slow
	// reflection-based checkInitializedSlow path (which calls Descriptor()).
	// All gogo-generated types use proto3 semantics and have no required fields.
	lm.methods.CheckInitialized = func(protoiface.CheckInitializedInput) (protoiface.CheckInitializedOutput, error) {
		return protoiface.CheckInitializedOutput{}, nil
	}
	return lm
}

// legacyMessage implements protoreflect.Message for gogo-generated types.
// The protoiface.Methods are embedded by value so that ProtoMethods() can return
// a pointer to the embedded struct without any additional heap allocation.
type legacyMessage struct {
	methods protoiface.Methods
	self    protoreflect.ProtoMessage
	valid   bool
}

var _ protoreflect.Message = (*legacyMessage)(nil)

// ProtoMethods returns the fast-path methods (Size, Marshal, Unmarshal,
// CheckInitialized). The returned pointer is stable for the lifetime of the
// legacyMessage; no allocation occurs on repeated calls.
func (lm *legacyMessage) ProtoMethods() *protoiface.Methods { return &lm.methods }

// Interface returns the original proto.Message wrapper passed to LegacyMessage.
func (lm *legacyMessage) Interface() protoreflect.ProtoMessage { return lm.self }

// IsValid reports whether the wrapped message is non-nil.
func (lm *legacyMessage) IsValid() bool { return lm.valid }

// GetUnknown and SetUnknown are no-ops; gogo types handle unknown fields internally.
func (lm *legacyMessage) GetUnknown() protoreflect.RawFields { return nil }
func (lm *legacyMessage) SetUnknown(protoreflect.RawFields)  {}

// WhichOneof always returns nil; v1alpha1 types have no oneof fields.
func (lm *legacyMessage) WhichOneof(protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	return nil
}

// The methods below require a compiled proto file descriptor.  They are never
// invoked by the gRPC binary codec or the grpc-gateway HTTP path, so they
// deliberately panic with an actionable message rather than silently returning
// zero/empty values that could mask bugs.

const errNeedDescriptor = "proto field reflection is not supported for gogo-generated v1alpha1 types: " +
	"no compiled file descriptor is available. " +
	"Use encoding/json or the gRPC binary codec instead."

func (lm *legacyMessage) Descriptor() protoreflect.MessageDescriptor {
	panic(errNeedDescriptor)
}
func (lm *legacyMessage) Type() protoreflect.MessageType { panic(errNeedDescriptor) }
func (lm *legacyMessage) New() protoreflect.Message      { panic(errNeedDescriptor) }
func (lm *legacyMessage) Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	panic(errNeedDescriptor)
}
func (lm *legacyMessage) Has(protoreflect.FieldDescriptor) bool { panic(errNeedDescriptor) }
func (lm *legacyMessage) Clear(protoreflect.FieldDescriptor)    { panic(errNeedDescriptor) }
func (lm *legacyMessage) Get(protoreflect.FieldDescriptor) protoreflect.Value {
	panic(errNeedDescriptor)
}
func (lm *legacyMessage) Set(protoreflect.FieldDescriptor, protoreflect.Value) {
	panic(errNeedDescriptor)
}
func (lm *legacyMessage) Mutable(protoreflect.FieldDescriptor) protoreflect.Value {
	panic(errNeedDescriptor)
}
func (lm *legacyMessage) NewField(protoreflect.FieldDescriptor) protoreflect.Value {
	panic(errNeedDescriptor)
}
