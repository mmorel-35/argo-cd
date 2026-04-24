//go:build ignore
// +build ignore

// Code generated to add proto.Message compatibility for grpc-gateway v2.
// This file is a vendor patch maintained in hack/vendor-patches/k8s.io/api/core/v1/
// and must be restored after running go mod vendor (see Makefile target apply-vendor-patches).
//
// These stubs allow gogo-generated k8s types to satisfy the google.golang.org/protobuf/proto.Message
// interface required by grpc-gateway v2 generated code at compile time.
// ProtoReflect() is never called at runtime: all HTTP responses are marshaled via encoding/json.

package v1

import "google.golang.org/protobuf/reflect/protoreflect"

func (*Event) ProtoReflect() protoreflect.Message     { return nil }
func (*EventList) ProtoReflect() protoreflect.Message { return nil }
func (*EventSeries) ProtoReflect() protoreflect.Message { return nil }
