Never use code comments.

Prefer the recipes in `justfile` for common development tasks. They run mostly Dockerized command and should be used instead of host-machine commands.

Use Ginkgo and Gomega for unit tests. See the ginkgo-overview skill for more.

Never put conditional branches in a `DescribeTable` body. If table rows require different branches, write separate specs.
