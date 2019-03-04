package state

//go:generate go run $GOPATH/src/github.com/oneiro-ndev/generator/cmd/nomsify $GOPATH/src/github.com/oneiro-ndev/metanode/pkg/meta/state
//go:generate find $GOPATH/src/github.com/oneiro-ndev/metanode/pkg/meta/state -name "*noms_gen*.go" -maxdepth 1 -exec goimports -w {} ;
