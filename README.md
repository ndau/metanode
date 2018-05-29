# metanode

In writing ABCI applications, there is some busywork involved in defining what exactly comprises a transaction, how transactions should interact, etc. It's perfectly fine to write it all out once, but we have three chains. Why not extract the common aspects?

That's what this repo is about: it's not a full ABCI application, but it contains common definitions which are useful for implementing ABCI applications.

## Dependencies

In go, libraries don't vendor their dependencies: it's simply not done. Attempting to do so invariably ends in tears.

Unfortunately, this library wants UUIDs, and there is no good UUID library for go right now:

- [`nu7hatch/go-uuid`](https://github.com/nu7hatch/gouuid) doesn't actually generate valid UUIDs
- [`google/uuid`](https://github.com/google/uuid) specifically warns that its API is currently unstable
- [`satori/go.uuid`](github.com/satori/go.uuid) has more stars than either on github, but the head of its `master` branch has a breaking API difference from its most recent tagged version. Its maintainer appears unwilling to create a v2 release, so to use this library requires that downstream dependencies specifically declare which tag to use.

We chose `satori/go.uuid` as the most common option, so we need downstream code to specify the required version. In the interest of future-compatibility, we choose the new API instead of the currently tagged API. This therefore requires that downstream dependencies of `metanode` specify that the correct version of the `go.uuid` library be `master`. In `glide.yaml`:

```yaml
- package: github.com/satori/go.uuid
  version: master
```
