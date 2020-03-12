package state

//go:generate go run $GOPATH/src/github.com/ndau/generator/cmd/nomsify $GOPATH/src/github.com/ndau/metanode/pkg/meta/state
//go:generate find $GOPATH/src/github.com/ndau/metanode/pkg/meta/state -name "*noms_gen*.go" -maxdepth 1 -exec goimports -w {} ;
// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


