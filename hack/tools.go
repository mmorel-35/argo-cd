//go:build tools
// +build tools

package tools

import (
	// k8s.io/code-generator is vendored to get generate-groups.sh, and k8s codegen utilities
	_ "k8s.io/code-generator"
	_ "k8s.io/code-generator/cmd/client-gen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "k8s.io/code-generator/cmd/defaulter-gen"
	_ "k8s.io/code-generator/cmd/go-to-protobuf"
	_ "k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo"
	_ "k8s.io/code-generator/cmd/informer-gen"
	_ "k8s.io/code-generator/cmd/lister-gen"

	// openapi-gen is vendored because upstream does not have tagged releases
	_ "k8s.io/kube-openapi/cmd/openapi-gen"
)
