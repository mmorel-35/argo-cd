# expecterlint Integration

This document describes the integration of expecterlint into the Argo CD project for better mock handling.

## What is expecterlint?

`expecterlint` is a linter that checks if calls to `.On("Method")` in testify mocks can be replaced with the more type-safe `.EXPECT().Method()` syntax introduced in mockery v2.10.0.

Benefits of using `.EXPECT()`:
- **Type safety**: Argument types are checked at compile time
- **Better IDE support**: Auto-completion and parameter hints
- **Cleaner code**: Type-safe callbacks instead of `mock.Arguments`

## Changes Made

1. **Tooling Integration**:
   - Added expecterlint to `hack/installers/install-lint-tools.sh`
   - Added Makefile targets: `lint-expecter-local` and `lint-expecter-fix-local`
   - Created helper script for building from source due to Go version compatibility

2. **Code Conversion**:
   - Converted `.On()` calls to `.EXPECT()` across test files
   - Updated `.Run()` callbacks to use type-safe function signatures
   - Changed `.Return(func1, func2)` patterns to `.RunAndReturn(func)`

3. **Documentation**:
   - Updated `docs/developer-guide/static-code-analysis.md` with expecterlint usage
   - Added examples of common conversion patterns

## Usage

### Check for issues (without modifying files):
```bash
make lint-expecter-local
```

### Auto-fix issues:
```bash
make lint-expecter-fix-local
```

### Manual fixes required

After running the auto-fix, you may need to manually update:

1. **`.Run()` callbacks** - Use type-safe signatures:
   ```go
   // Before
   mock.On("Method", mock.Anything).Run(func(args mock.Arguments) {
       param := args.Get(0).(SomeType)
   })

   // After
   mock.EXPECT().Method(mock.Anything).Run(func(param SomeType) {
       // Use param directly
   })
   ```

2. **Dynamic return values** - Use `.RunAndReturn()`:
   ```go
   // Before
   mock.On("Method", mock.Anything).Return(
       func(arg SomeType) ReturnType { return compute(arg) },
       func(arg SomeType) error { return computeError(arg) },
   )

   // After
   mock.EXPECT().Method(mock.Anything).RunAndReturn(
       func(arg SomeType) (ReturnType, error) {
           return compute(arg), computeError(arg)
       },
   )
   ```

## Known Issues

### Go Version Compatibility

The released version of expecterlint (v1.1.0) was built with Go 1.24, but this repository requires Go 1.25. If you encounter version compatibility errors, build from source:

```bash
bash hack/installers/install-expecterlint-from-source.sh
```

This will clone the expecterlint repository, update the Go version in go.mod, and build/install it locally.

## Testing

All tests continue to pass after the conversion. You can verify by running:
```bash
go test -short ./...
```

## References

- [expecterlint GitHub](https://github.com/d0ubletr0uble/expecterlint)
- [Mockery Expecter Structs](https://vektra.github.io/mockery/v3.5/template/testify/#expecter-structs)
