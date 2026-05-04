// Package v1alpha1 serialisation bridge.
//
// This file wires the gogo-generated v1alpha1 types into the
// google.golang.org/protobuf (APIv2) world via util/proto.LegacyMessage.
// See that package's documentation for the detailed rationale and for the
// explanation of why protoadapt.MessageV2Of cannot be used here.

package v1alpha1

import (
protoutil "github.com/argoproj/argo-cd/v3/util/proto"
"google.golang.org/protobuf/reflect/protoreflect"
)

// legacyReflect returns a protoreflect.Message for any v1alpha1 type that
// satisfies both protoreflect.ProtoMessage (for Interface()) and
// protoutil.Serializable (for fast-path Marshal/Unmarshal/Size).
//
// It is called from every ProtoReflect() stub in proto_reflect_compat.go.
func legacyReflect(self interface {
protoreflect.ProtoMessage
protoutil.Serializable
}) protoreflect.Message {
return protoutil.LegacyMessage(self, self)
}
