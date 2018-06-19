# `meta.app`: Generic ABCI application implementation

All of Oneiro's blockchains are implemented in terms of Tendermint's ABCI interface specification. There's actually a fair amount of busywork involved in satisfying that interface.

The purpose of this package is to abstract that interface's implementation.

## Usage Checklist

- [ ] subclass the meta-application:

    ```go
    import meta "github.com/oneiro-ndev/metanode/pkg/meta.app"
    type MyApp struct {
        *meta.App
        ...
    }
    ```

- [ ] implement backing state conforming to the `State` interface.
- [ ] implement some number of transactions conforming to the `metatx.Transactable` interface
- [ ] construct a `metatx.TxIDMap` enumerating all valid `Transactable`s
- [ ] initialize your application as follows:

    ```go
    func NewApp(...) (*MyApp, error) {
        metaapp, err := meta.NewApp(dbSpec, name, new(MyState), TxIDs)
        if err != nil {
            return nil, errors.Wrap(err, "NewApp failed to create metaapp")
        }

        // init your app's fields here

        app := App{
            metaapp,
            ...
        }
        app.App.SetChild(&app)
        return &app, nil
    }
    ```
