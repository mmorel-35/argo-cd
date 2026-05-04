package proto_test

import (
	"testing"

	protoutil "github.com/argoproj/argo-cd/v3/util/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

// fakeMsg is a minimal gogo-generated type for testing.
type fakeMsg struct {
	data []byte
}

func (f *fakeMsg) Reset()                      { *f = fakeMsg{} }
func (f *fakeMsg) String() string              { return string(f.data) }
func (*fakeMsg) ProtoMessage()                 {}
func (f *fakeMsg) ProtoReflect() protoreflect.Message { return protoutil.LegacyMessage(f, f) }
func (f *fakeMsg) Marshal() ([]byte, error)    { return f.data, nil }
func (f *fakeMsg) Unmarshal(b []byte) error    { f.data = b; return nil }
func (f *fakeMsg) Size() int                   { return len(f.data) }

func TestLegacyMessageInterface(t *testing.T) {
	orig := &fakeMsg{data: []byte("hello")}
	m := protoutil.LegacyMessage(orig, orig)

	if !m.IsValid() {
		t.Error("IsValid: want true, got false")
	}
	if m.Interface() != orig {
		t.Error("Interface: did not return original value")
	}
	if m.GetUnknown() != nil {
		t.Error("GetUnknown: want nil")
	}
	if m.WhichOneof(nil) != nil {
		t.Error("WhichOneof: want nil")
	}
	m.SetUnknown(nil) // must not panic
}

func TestLegacyMessageProtoMethods(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	orig := &fakeMsg{data: payload}
	m := protoutil.LegacyMessage(orig, orig)

	methods := m.ProtoMethods()
	if methods == nil {
		t.Fatal("ProtoMethods: nil")
	}
	// Size
	sizeOut := methods.Size(protoiface.SizeInput{Message: m})
	if sizeOut.Size != len(payload) {
		t.Errorf("Size: want %d, got %d", len(payload), sizeOut.Size)
	}
	// Marshal
	out, err := methods.Marshal(protoiface.MarshalInput{Message: m})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(out.Buf) != string(payload) {
		t.Errorf("Marshal: want %v, got %v", payload, out.Buf)
	}
	// Unmarshal
	dest := &fakeMsg{}
	destM := protoutil.LegacyMessage(dest, dest)
	_, err = destM.ProtoMethods().Unmarshal(protoiface.UnmarshalInput{
		Message: destM,
		Buf:     []byte{0x04, 0x05},
	})
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(dest.data) != "\x04\x05" {
		t.Errorf("Unmarshal data: got %v", dest.data)
	}
}

func TestLegacyMessageProtoMethodsStable(t *testing.T) {
	// ProtoMethods() must return a stable pointer (no allocation per call).
	orig := &fakeMsg{data: []byte("x")}
	m := protoutil.LegacyMessage(orig, orig)
	p1 := m.ProtoMethods()
	p2 := m.ProtoMethods()
	if p1 != p2 {
		t.Error("ProtoMethods: returned different pointers on successive calls")
	}
}

func TestLegacyMessageNilIsValid(t *testing.T) {
	m := protoutil.LegacyMessage(nil, &fakeMsg{})
	if m.IsValid() {
		t.Error("IsValid: want false for nil self")
	}
}

func TestLegacyMessageReflectionPanics(t *testing.T) {
	// Descriptor and field-reflection methods must panic, not silently return zero values.
	orig := &fakeMsg{}
	m := protoutil.LegacyMessage(orig, orig)

	mustPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s: expected panic, got none", name)
			}
		}()
		fn()
	}

	mustPanic("Descriptor", func() { _ = m.Descriptor() })
	mustPanic("Type", func() { _ = m.Type() })
	mustPanic("New", func() { _ = m.New() })
	mustPanic("Range", func() { m.Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool { return true }) })
	mustPanic("Has", func() { _ = m.Has(nil) })
	mustPanic("Clear", func() { m.Clear(nil) })
	mustPanic("Get", func() { _ = m.Get(nil) })
	mustPanic("Set", func() { m.Set(nil, protoreflect.Value{}) })
	mustPanic("Mutable", func() { _ = m.Mutable(nil) })
	mustPanic("NewField", func() { _ = m.NewField(nil) })
}

// realProtoMsg is a fakeMsg that round-trips via proto.Marshal / proto.Unmarshal.
func TestProtoMarshalRoundTrip(t *testing.T) {
	// Encode some raw bytes via a fakeMsg.
	src := &fakeMsg{data: []byte{0x0a, 0x05, 'h', 'e', 'l', 'l', 'o'}}
	b, err := proto.Marshal(src)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("proto.Marshal: empty output")
	}
	dst := &fakeMsg{}
	if err := proto.Unmarshal(b, dst); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}
	if string(dst.data) != string(src.data) {
		t.Errorf("round-trip: want %v, got %v", src.data, dst.data)
	}
}
