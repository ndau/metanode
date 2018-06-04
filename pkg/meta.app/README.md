# `meta.app`: Generic ABCI application implementation

All of Oneiro's blockchains are implemented in terms of Tendermint's ABCI interface specification. There's actually a fair amount of busywork involved in satisfying that interface.

The purpose of this package is to abstract that interface's implementation.

## Usage Checklist

- [ ] subclass the meta-application:

    ```go
    import meta "github.com/oneir-ndev/metanode/pkg/meta.app"
    type MyApp struct {
        *meta.App
        ...
    }
    ```

- [ ] implement backing state conforming to the `State` interface.
- [ ] implement some number of transactions conforming to the `metatx.Transactable` interface
- [ ] override the `app.GetTxIDs` method to return an appropriate map
- [ ] initialize your application as follows:

    ```go
    func (m *MyApp) NewApp(...) (*MyApp, error) {
        ...
        app := &MyApp{
            meta.NewApp(spec, name, state),
            ...
        }
        return app, nil
    }
    ```
