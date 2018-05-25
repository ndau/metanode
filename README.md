# metanode

In writing ABCI applications, there is some busywork involved in defining what exactly comprises a transaction, how transactions should interact, etc. It's perfectly fine to write it all out once, but we have three chains. Why not extract the common aspects?

That's what this repo is about: it's not a full ABCI application, but it contains common definitions which are useful for implementing ABCI applications.
