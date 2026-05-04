package v1alpha1

import (
	"testing"

	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto" // register proto codec
	"google.golang.org/protobuf/proto"
)

// TestProtoMarshalRoundTrip verifies that v1alpha1 types can be serialised and
// deserialised by both the proto package and the gRPC binary codec without
// panicking or corrupting data.
func TestProtoMarshalRoundTrip(t *testing.T) {
	app := &Application{}
	app.Name = "test-app"
	app.Namespace = "default"
	app.Spec.Destination.Server = "https://kubernetes.default.svc"
	app.Spec.Destination.Namespace = "prod"
	app.Spec.Source = &ApplicationSource{
		RepoURL:        "https://github.com/example/repo",
		Path:           "charts/myapp",
		TargetRevision: "HEAD",
	}
	app.Status.Health.Status = "Healthy"

	// proto.Marshal / proto.Unmarshal round-trip.
	b, err := proto.Marshal(app)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("proto.Marshal: empty output")
	}
	app2 := &Application{}
	if err := proto.Unmarshal(b, app2); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}
	if app2.Name != app.Name {
		t.Errorf("Name: want %q, got %q", app.Name, app2.Name)
	}
	if app2.Spec.Destination.Server != app.Spec.Destination.Server {
		t.Errorf("Destination.Server: want %q, got %q", app.Spec.Destination.Server, app2.Spec.Destination.Server)
	}
	if app2.Status.Health.Status != app.Status.Health.Status {
		t.Errorf("Health.Status: want %q, got %q", app.Status.Health.Status, app2.Status.Health.Status)
	}
}

// TestProtoGRPCCodecRoundTrip verifies that the gRPC binary codec can serialise
// and deserialise v1alpha1 types, which is the primary production code path.
func TestProtoGRPCCodecRoundTrip(t *testing.T) {
	codec := encoding.GetCodecV2("proto")
	if codec == nil {
		t.Skip("proto codec not registered")
	}

	proj := &AppProject{}
	proj.Name = "my-project"
	proj.Namespace = "argocd"
	proj.Spec.Description = "test project"
	proj.Spec.SourceRepos = []string{"https://github.com/example/repo"}

	bufs, err := codec.Marshal(proj)
	if err != nil {
		t.Fatalf("codec.Marshal: %v", err)
	}

	proj2 := &AppProject{}
	if err := codec.Unmarshal(bufs, proj2); err != nil {
		t.Fatalf("codec.Unmarshal: %v", err)
	}
	if proj2.Name != proj.Name {
		t.Errorf("Name: want %q, got %q", proj.Name, proj2.Name)
	}
	if proj2.Spec.Description != proj.Spec.Description {
		t.Errorf("Spec.Description: want %q, got %q", proj.Spec.Description, proj2.Spec.Description)
	}
	if len(proj2.Spec.SourceRepos) != 1 || proj2.Spec.SourceRepos[0] != proj.Spec.SourceRepos[0] {
		t.Errorf("Spec.SourceRepos: want %v, got %v", proj.Spec.SourceRepos, proj2.Spec.SourceRepos)
	}
}

// TestProtoClusterRoundTrip exercises a type with nested structs and verifies
// that the ProtoReflect / ProtoMethods plumbing works for diverse field types.
func TestProtoClusterRoundTrip(t *testing.T) {
	cluster := &Cluster{
		Name:   "my-cluster",
		Server: "https://my-cluster:6443",
		Config: ClusterConfig{
			BearerToken: "supersecret",
			TLSClientConfig: TLSClientConfig{
				Insecure: true,
			},
		},
	}

	b, err := proto.Marshal(cluster)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	cluster2 := &Cluster{}
	if err := proto.Unmarshal(b, cluster2); err != nil {
		t.Fatalf("proto.Unmarshal: %v", err)
	}
	if cluster2.Server != cluster.Server {
		t.Errorf("Server: want %q, got %q", cluster.Server, cluster2.Server)
	}
	if cluster2.Config.BearerToken != cluster.Config.BearerToken {
		t.Errorf("BearerToken: want %q, got %q", cluster.Config.BearerToken, cluster2.Config.BearerToken)
	}
	if cluster2.Config.TLSClientConfig.Insecure != cluster.Config.TLSClientConfig.Insecure {
		t.Errorf("Insecure: want %v, got %v", cluster.Config.TLSClientConfig.Insecure, cluster2.Config.TLSClientConfig.Insecure)
	}
}

// TestProtoReflectIsValid verifies that ProtoReflect() returns a valid message
// for non-nil types and an invalid message for nil pointers.
func TestProtoReflectIsValid(t *testing.T) {
	app := &Application{}
	if !app.ProtoReflect().IsValid() {
		t.Error("non-nil Application: ProtoReflect().IsValid() want true")
	}
	var nilApp *Application
	if nilApp.ProtoReflect().IsValid() {
		t.Error("nil Application: ProtoReflect().IsValid() want false")
	}
}
