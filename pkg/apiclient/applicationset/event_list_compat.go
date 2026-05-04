package applicationset

import (
	protoutil "github.com/argoproj/argo-cd/v3/util/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	corev1 "k8s.io/api/core/v1"
)

// eventListMessage wraps *corev1.EventList so it satisfies the proto.Message interface
// required by grpc-gateway v2 generated code at compile time.
//
// The gRPC binary codec and grpc-gateway HTTP path both use only the fast-path
// Marshal/Unmarshal/Size methods provided by LegacyMessage. HTTP responses are
// serialised via encoding/json (our custom JSONMarshaler), so the binary codec
// methods are only used on the gRPC path.
type eventListMessage struct {
	*corev1.EventList
}

func (e *eventListMessage) ProtoReflect() protoreflect.Message {
	return protoutil.LegacyMessage(e, e.EventList)
}
