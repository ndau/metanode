# metanode

In writing ABCI applications, there is some busywork involved in defining what exactly comprises a transaction, how transactions should interact, etc. It's perfectly fine to write it all out once, but we have three chains. Why not extract the common aspects?

That's what this repo is about: it's not a full ABCI application, but it contains common definitions which are useful for implementing ABCI applications.

## Dependencies

In go, libraries don't vendor their dependencies: it's simply not done. Attempting to do so invariably ends in tears.

We're currently using github.com/gofrs/uuid as the UUID library, which seems to be the one the go community has
settled on. We use V1 UUIDs but at least v3.2 of the library.
