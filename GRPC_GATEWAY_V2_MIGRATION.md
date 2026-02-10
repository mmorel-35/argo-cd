# gRPC-Gateway v2 Migration Analysis

## Executive Summary

A complete migration from grpc-gateway v1 to v2 is **blocked** by a fundamental incompatibility between:
- grpc-gateway v2 (requires `google.golang.org/protobuf`)
- Kubernetes code generators (generate `gogo/protobuf` code)

## Technical Background

### The Incompatibility

gRPC-Gateway v2 requires all protobuf types to implement `protoreflect.ProtoMessage` from `google.golang.org/protobuf`. However, Argo CD's CRD types in `pkg/apis/application/v1alpha1` are generated using Kubernetes' `go-to-protobuf` tool, which outputs `gogo/protobuf` compatible code that does NOT implement this interface.

###Impact Scope

**Proto files referencing v1alpha1 types (10 out of 18 total):**
- server/application/application.proto
- server/applicationset/applicationset.proto
- server/certificate/certificate.proto
- server/cluster/cluster.proto
- server/gpgkey/gpgkey.proto
- server/project/project.proto
- server/repocreds/repocreds.proto
- server/repository/repository.proto
- server/settings/settings.proto
- reposerver/repository/repository.proto

**Result:** Cannot do partial migration - almost all services use v1alpha1 types.

## Work Completed

✅ Updated dependencies (grpc-gateway/v2 v2.27.3, googleapis)
✅ Updated tooling (protoc-gen-grpc-gateway v2, protoc-gen-openapiv2)
✅ Updated proto generation scripts
✅ Fixed proto file conflicts (removed HTTP annotations from deprecated RPCs)
✅ Successfully generated all proto files with v2 tooling

❌ Build fails due to type incompatibility

## Migration Options

### Option 1: Stay on grpc-gateway v1 ⚠️ (Current State)
**Pros:**
- No code changes required
- No breaking changes
- Immediate solution

**Cons:**
- v1.16.0 is deprecated (last release 2020)
- No security updates
- Technical debt accumulation

### Option 2: Full Migration to google.golang.org/protobuf ✅ (Recommended)
**Requires:**
1. Migrate Kubernetes CRD generation from gogo/protobuf to google.golang.org/protobuf
2. Replace `go-to-protobuf` with modern alternative OR manually maintain proto files
3. Regenerate all v1alpha1 types with `protoc-gen-go`
4. Refactor business logic methods (~1000+ lines across 50+ types)
5. Update all code using gogo-specific features
6. Extensive testing for CRD compatibility

**Estimated Effort:** 2-4 weeks

**Pros:**
- Future-proof solution
- Active maintenance and security updates
- Better OpenAPI v2 support
- Aligns with upstream Kubernetes direction

**Cons:**
- Significant development effort
- Potential breaking changes in CRD structure
- Risk of bugs during transition
- Requires careful testing

### Option 3: Hybrid with Compatibility Bridge 🔧 (Not Recommended)
**Approach:**
- Keep v1alpha1 on gogo/protobuf
- Create adapter layer between gogo and google protobuf types
- Use reflection/code generation for conversion

**Pros:**
- Incremental migration possible

**Cons:**
- High complexity and maintenance burden
- Performance overhead
- Still technical debt
- May not cover all edge cases

## Recommendation

**Option 2 (Full Migration)** is the proper long-term solution, BUT requires:
- Commitment to address technical debt
- Dedicated time for development and testing
- Acceptance of potential breaking changes

**Interim Solution:**
If immediate v2 migration is not feasible:
1. Revert to grpc-gateway v1.16.0
2. Add dependency monitoring for security issues
3. Plan Option 2 migration for next quarter

## Next Steps

**Decision Required:**
1. Accept Option 2 scope and proceed with full migration?
2. OR defer migration and revert to v1 with planned timeline?

**If proceeding with Option 2:**
1. Investigate Kubernetes code generation alternatives
2. Create proof-of-concept for v1alpha1 migration
3. Plan phased rollout with extensive testing
4. Document all breaking changes

## References

- Issue #92: First migration attempt (buf integration)
- Issue #93: Second migration attempt (identified gogo/protobuf blocker)
- [grpc-gateway v2 migration guide](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/migrating_from_v1/)
- [Kubernetes go-to-protobuf documentation](https://github.com/kubernetes/code-generator)

## Files Modified in This PR

### Tooling & Dependencies
- `go.mod` - Updated to grpc-gateway/v2, added googleapis
- `hack/tools.go` - Updated imports to v2
- `hack/installers/install-codegen-go-tools.sh` - Install v2 plugins
- `hack/generate-proto.sh` - Use v2 plugins and googleapis

### Proto Files
- `server/repository/repository.proto` - Removed HTTP annotations from deprecated RPCs

### Generated Files (v2 compatible, but won't compile)
- All `*.pb.gw.go` files (20 files) - Regenerated with v2
- `assets/swagger.json` - Regenerated with OpenAPI v2
- `pkg/apiclient/repository/repository.pb.go` - Regenerated

---

**Status:** ⏸️ Blocked pending decision on migration approach
**Last Updated:** 2026-02-10
