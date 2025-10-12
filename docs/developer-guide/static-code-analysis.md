# Static Code Analysis

We use the following static code analysis tools:

* `golangci-lint` and `eslint` for compile time linting
* `expecterlint` for checking if mock.On() can be replaced with mock.EXPECT() for better type safety
* [codecov.io](https://codecov.io/gh/argoproj/argo-cd) - for code coverage
* [snyk.io](https://app.snyk.io/org/argoproj/projects) - for image scanning
* [sonarcloud.io](https://sonarcloud.io/organizations/argoproj/projects) - for code scans and security alerts

These are at least run daily or on each pull request.

## expecterlint

Since `v2.10.0` [mockery](https://github.com/vektra/mockery) introduced [Expecter Structs](https://vektra.github.io/mockery/v3.5/template/testify/#expecter-structs). The `expecterlint` tool checks if calls to `.On("Method")` could be replaced with the more type-safe `.EXPECT().Method()` syntax.

### Usage

To check for potential improvements without modifying files:
```bash
make lint-expecter-local
```

To automatically convert `.On()` calls to `.EXPECT()`:
```bash
make lint-expecter-fix-local
```

**Note:** When converting `.On()` to `.EXPECT()`, you may need to manually update `.Run()` callbacks to use type-safe function signatures instead of `mock.Arguments`. For example:

```go
// Before (with .On())
mock.On("Method", mock.Anything).Run(func(args mock.Arguments) {
    param := args.Get(0).(SomeType)
})

// After (with .EXPECT())
mock.EXPECT().Method(mock.Anything).Run(func(param SomeType) {
    // Use param directly
})
```

Additionally, when using functions that return multiple values dynamically, use `.RunAndReturn()` instead of `.Return()`:

```go
// Before (with .On())
mock.On("Method", mock.Anything).Return(
    func(arg SomeType) ReturnType { return computeReturn(arg) },
    func(arg SomeType) error { return computeError(arg) },
)

// After (with .EXPECT())
mock.EXPECT().Method(mock.Anything).RunAndReturn(
    func(arg SomeType) (ReturnType, error) {
        return computeReturn(arg), computeError(arg)
    },
)
```
