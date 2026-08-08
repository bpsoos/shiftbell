Never use code comments.

## general commands

Prefer the recipes in `justfile` for common development tasks. They run mostly Dockerized command and should be used instead of host-machine commands.

## code conventions

Constructors assign provided configuration and dependencies without defaults or guard clauses unless a specific use case requires them.

## testing

Use Ginkgo and Gomega for unit tests. See the ginkgo-overview skill for more.

Testing strategy per layer:

- Endpoint unit tests exercise public endpoint methods through Echo and `httptest`, with dependencies mocked.
- Do not write unit tests for the view layer.
- Service unit tests have dependencies mocked.
- Persistence integration-unit tests use real SQLite.
Keep each test beside its implementation and name it `<implementation>_test.go`. Do not add cross-cutting test files testing multiple components. No in-package or private method testing.
Never put conditional branches in a `DescribeTable` body. If table rows require different branches, write separate specs.

## styles, css

Prefer colocated, flattened component CSS using `cssScopeInline`; keep only truly shared styles global. Keep scoped `<style>` parents free of stable IDs so HTMX swaps retain their scope classes.
