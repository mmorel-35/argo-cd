package applicationset

import (
	"errors"

	"github.com/argoproj/pkg/v2/grpc/http"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	corev1 "k8s.io/api/core/v1"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
)

// k8sEventListWrapper wraps *corev1.EventList to satisfy google.golang.org/protobuf/proto.Message
// for grpc-gateway v2. k8s v0.35+ moved ProtoMessage() behind a build tag, so EventList no longer
// implements the new proto.Message interface. ProtoReflect() is never called at runtime because
// all HTTP responses are marshaled via encoding/json by our custom Marshaler.
type k8sEventListWrapper struct {
	*corev1.EventList
}

func (w *k8sEventListWrapper) ProtoReflect() protoreflect.Message { return nil }

func init() {
	forward_ApplicationSetService_Watch_0 = http.NewStreamForwarder(func(message proto.Message) (string, error) {
		event, ok := message.(*v1alpha1.ApplicationSetWatchEvent)
		if !ok {
			return "", errors.New("unexpected message type")
		}
		return event.ApplicationSet.Name, nil
	})
}
