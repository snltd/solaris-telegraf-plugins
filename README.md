# SmartOS and Solaris Telegraf Plugins

A collection of Telegraf plugins to collect data on a SmartOS or
Solaris system, from the global or local zone.


## Building

Get the Telegraf source, and in `plugins/input/all/all.go` add a
bunch of lines line

```go
_ "github.com/snltd/solaris-telegraf-plugins/smf_svc"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_io"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_memory"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_network"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_nfs_client"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_nfs_server"
```

Then build Telegraf as normal.

## Caveats

I'm learning Go as I write these. They are badly written.

## Contributing

Fork it, fix it, push it, PR it.
