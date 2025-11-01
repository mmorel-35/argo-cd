#! /usr/bin/env bash

# This script auto-generates protobuf related files. It is intended to be run manually when either
# API types are added/modified, or server gRPC calls are added. The generated files should then
# be checked into source control.

set -x
set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC2128
PROJECT_ROOT=$(
    cd "$(dirname "${BASH_SOURCE}")"/..
    pwd
)
PATH="${PROJECT_ROOT}/dist:${PATH}"
GOPATH=$(go env GOPATH)
GOPATH_PROJECT_ROOT="${GOPATH}/src/github.com/argoproj/argo-cd"

# output tool versions
go version
protoc --version
buf --version
swagger version
jq --version

export GO111MODULE=off

# Generate pkg/apis/<group>/<apiversion>/(generated.proto,generated.pb.go)
# NOTE: any dependencies of our types to the k8s.io apimachinery types should be added to the
# --apimachinery-packages= option so that go-to-protobuf can locate the types, but prefixed with a
# '-' so that go-to-protobuf will not generate .proto files for it.
PACKAGES=(
    github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1
)
APIMACHINERY_PKGS=(
    +k8s.io/apimachinery/pkg/util/intstr
    +k8s.io/apimachinery/pkg/api/resource
    +k8s.io/apimachinery/pkg/runtime/schema
    +k8s.io/apimachinery/pkg/runtime
    k8s.io/apimachinery/pkg/apis/meta/v1
    k8s.io/api/core/v1
    k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1
)

export GO111MODULE=on
[ -e ./v3 ] || ln -s . v3
[ -e "${GOPATH_PROJECT_ROOT}" ] || (mkdir -p "$(dirname "${GOPATH_PROJECT_ROOT}")" && ln -s "${PROJECT_ROOT}" "${GOPATH_PROJECT_ROOT}")

# protoc_include is the include directory containing the .proto files distributed with protoc binary
if [ -d /dist/protoc-include ]; then
    # containerized codegen build
    protoc_include=/dist/protoc-include
else
    # local codegen build
    protoc_include=${PROJECT_ROOT}/dist/protoc-include
fi

# go-to-protobuf expects dependency proto files to be in $GOPATH/src. Copy them there.
rm -rf "${GOPATH}/src/github.com/gogo/protobuf" && mkdir -p "${GOPATH}/src/github.com/gogo" && cp -r "${PROJECT_ROOT}/vendor/github.com/gogo/protobuf" "${GOPATH}/src/github.com/gogo"
rm -rf "${GOPATH}/src/k8s.io/apimachinery" && mkdir -p "${GOPATH}/src/k8s.io" && cp -r "${PROJECT_ROOT}/vendor/k8s.io/apimachinery" "${GOPATH}/src/k8s.io"
rm -rf "${GOPATH}/src/k8s.io/api" && mkdir -p "${GOPATH}/src/k8s.io" && cp -r "${PROJECT_ROOT}/vendor/k8s.io/api" "${GOPATH}/src/k8s.io"
rm -rf "${GOPATH}/src/k8s.io/apiextensions-apiserver" && mkdir -p "${GOPATH}/src/k8s.io" && cp -r "${PROJECT_ROOT}/vendor/k8s.io/apiextensions-apiserver" "${GOPATH}/src/k8s.io"

go-to-protobuf \
    --go-header-file="${PROJECT_ROOT}"/hack/custom-boilerplate.go.txt \
    --packages="$(
        IFS=,
        echo "${PACKAGES[*]}"
    )" \
    --apimachinery-packages="$(
        IFS=,
        echo "${APIMACHINERY_PKGS[*]}"
    )" \
    --proto-import="${PROJECT_ROOT}"/vendor \
    --proto-import="${protoc_include}" \
    --output-dir="${GOPATH}/src/"

# go-to-protobuf modifies vendored code. Re-vendor code so it's available for subsequent steps.
go mod vendor

# Use buf to generate code from proto files instead of manual protoc loop.
# buf automatically discovers proto files based on buf.yaml configuration and uses
# the plugins defined in buf.gen.yaml (gogofast, grpc-gateway, swagger).
# With paths=source_relative, files are generated next to proto files, so we move them
# to their expected locations based on go_package declarations.
buf generate

# Move generated files from proto locations to their expected go_package locations
# server/* proto files have go_package pointing to pkg/apiclient/*
for dir in account application applicationset certificate cluster gpgkey notification project repocreds repository session settings version; do
    if [ -f "server/${dir}/${dir}.pb.go" ]; then
        mkdir -p "pkg/apiclient/${dir}"
        mv "server/${dir}/${dir}.pb.go" "pkg/apiclient/${dir}/"
    fi
    if [ -f "server/${dir}/${dir}.pb.gw.go" ]; then
        mkdir -p "pkg/apiclient/${dir}"
        mv "server/${dir}/${dir}.pb.gw.go" "pkg/apiclient/${dir}/"
    fi
done

# server/settings/oidc has no gateway file and stays in place
# (generated pb.go file location matches go_package)

# reposerver/repository -> reposerver/apiclient
if [ -f "reposerver/repository/repository.pb.go" ]; then
    mkdir -p "reposerver/apiclient"
    mv "reposerver/repository/repository.pb.go" "reposerver/apiclient/"
fi

# cmpserver/plugin -> cmpserver/apiclient
if [ -f "cmpserver/plugin/plugin.pb.go" ]; then
    mkdir -p "cmpserver/apiclient"
    mv "cmpserver/plugin/plugin.pb.go" "cmpserver/apiclient/"
fi

# commitserver/commit -> commitserver/apiclient
if [ -f "commitserver/commit/commit.pb.go" ]; then
    mkdir -p "commitserver/apiclient"
    mv "commitserver/commit/commit.pb.go" "commitserver/apiclient/"
fi

# util/askpass stays in place (matches go_package)

# This file is generated but should not be checked in.
rm -f util/askpass/askpass.swagger.json

[ -L "${GOPATH_PROJECT_ROOT}" ] && rm -rf "${GOPATH_PROJECT_ROOT}"
[ -L ./v3 ] && rm -rf v3

# collect_swagger gathers swagger files into a subdirectory
collect_swagger() {
    SWAGGER_ROOT="$1"
    SWAGGER_OUT="${PROJECT_ROOT}/assets/swagger.json"
    PRIMARY_SWAGGER=$(mktemp)
    COMBINED_SWAGGER=$(mktemp)

    cat <<EOF >"${PRIMARY_SWAGGER}"
{
  "swagger": "2.0",
  "info": {
    "title": "Consolidate Services",
    "description": "Description of all APIs",
    "version": "version not set"
  },
  "paths": {}
}
EOF

    rm -f "${SWAGGER_OUT}"

    find "${SWAGGER_ROOT}" -name '*.swagger.json' -exec swagger mixin --ignore-conflicts "${PRIMARY_SWAGGER}" '{}' \+ >"${COMBINED_SWAGGER}"
    jq -r 'del(.definitions[].properties[]? | select(."$ref"!=null and .description!=null).description) | del(.definitions[].properties[]? | select(."$ref"!=null and .title!=null).title) |
      # The "array" and "map" fields have custom unmarshaling. Modify the swagger to reflect this.
      .definitions.v1alpha1ApplicationSourcePluginParameter.properties.array = {"description":"Array is the value of an array type parameter.","type":"array","items":{"type":"string"}} |
      del(.definitions.v1alpha1OptionalArray) |
      .definitions.v1alpha1ApplicationSourcePluginParameter.properties.map = {"description":"Map is the value of a map type parameter.","type":"object","additionalProperties":{"type":"string"}} |
      del(.definitions.v1alpha1OptionalMap) |
      # Output for int64 is incorrect, because it is based on proto definitions, where int64 is a string. In our JSON API, we expect int64 to be an integer. https://github.com/grpc-ecosystem/grpc-gateway/issues/219
      (.definitions[]?.properties[]? | select(.type == "string" and .format == "int64")) |= (.type = "integer")
    ' "${COMBINED_SWAGGER}" |
        jq '.definitions.v1Time.type = "string" | .definitions.v1Time.format = "date-time" | del(.definitions.v1Time.properties)' |
        jq '.definitions.v1alpha1ResourceNode.allOf = [{"$ref": "#/definitions/v1alpha1ResourceRef"}] | del(.definitions.v1alpha1ResourceNode.properties.resourceRef) ' \
            >"${SWAGGER_OUT}"

    /bin/rm "${PRIMARY_SWAGGER}" "${COMBINED_SWAGGER}"
}

# clean up generated swagger files (should come after collect_swagger)
clean_swagger() {
    SWAGGER_ROOT="$1"
    find "${SWAGGER_ROOT}" -name '*.swagger.json' -delete
}

collect_swagger server
clean_swagger server
clean_swagger reposerver
clean_swagger controller
clean_swagger cmpserver
clean_swagger commitserver
