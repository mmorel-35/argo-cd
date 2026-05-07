# gRPC-Gateway v2 Migration Playbook

This guide defines the progressive migration strategy to move Argo CD from grpc-gateway v1/gogo-oriented patterns to grpc-gateway v2 with small, reviewable pull requests.

> [!NOTE]
> By default, migrations must maintain backward compatibility for external APIs unless maintainers explicitly approve a breaking change.

> [!WARNING]
> Do not combine Kubernetes dependency upgrades with grpc-gateway migration phases.

## Migration phases (stacked PRs)

### Phase A — Preflight blockers only
- Remove duplicate HTTP route annotations on deprecated RPCs.
- Land only hand-written blocker fixes needed for code generation.

### Phase B — EventList compatibility contract
- Decide and land one compatibility direction independently:
  - non-breaking wrapper/shim, or
  - breaking API-shape migration.
- Keep this phase isolated from broad regeneration.

### Phase C — Toolchain and generation pipeline
- Update codegen tooling/scripts only.
- Ensure deterministic generation workflow before bulk file churn.

### Phase D — Runtime wiring migration
- Update runtime imports/wiring in server/marshaler/forwarders.
- Keep manual logic changes minimal and explicit.

### Phase E — Generated gateway files batch
- Regenerate and review only `*.pb.gw.go` (plus minimal required compat glue).

### Phase F — Generated protobuf/grpc files batch
- Regenerate and review only `*.pb.go` and `*_grpc.pb.go`.
- Avoid manual logic edits in this phase.

### Phase G — First-party gogo cleanup and dependency pruning
- Remove obsolete first-party gogo code paths/imports.
- Finalize dependency cleanup and vendor normalization.

## Reviewability controls (all phases)

- Keep every PR single-purpose with an explicit **out of scope** section.
- Separate manual and generated changes in different commits.
- Include a short “why this PR exists in the sequence” section.
- Cap blast radius and split again if review scope becomes too large.
- Keep a tracker issue with a checklist for phases A→G.

## Validation gates by phase

- After phases A/B/C: codegen must succeed in CI.
- After phase D: API server startup, gateway route registration, and marshalling checks must pass.
- After phases E/F: generated-code consistency and targeted API smoke tests must pass.
- After phase G: no first-party gogo usage remains and dependency graph is stable.

## Risk and rollback rules

- Preserve non-breaking behavior unless an approved issue explicitly allows otherwise.
- If a phase destabilizes CI, revert only that phase.
- Do not include unrelated refactors in migration phases.

## Maintainer-facing PR narrative template

Each migration PR should include:
- Problem solved in this phase
- Why this phase is separate
- What remains (next phases)
- Validation evidence
- Explicit out-of-scope items
