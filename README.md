# SmartOS and Solaris Telegraf Plugins

A collection of Telegraf input plugins to collect data on a SmartOS or Solaris system, from the
global or local zone.

## The Plugins

## solaris_zones

Reports the number of zones in all of the states a zone can be in. `installed`, `running`, and
so-on.


## Building

Get the Telegraf source, and in `plugins/input/all/all.go` add a
bunch of lines line

```go
_ "github.com/snltd/solaris-telegraf-plugins/solaris_io"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_memory"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_network"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_nfs_client"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_nfs_server"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_smf"
_ "github.com/snltd/solaris-telegraf-plugins/solaris_zpool"
```

Then build Telegraf as normal.

## Caveats

I'm learning Go as I write these. They are badly written.

## Contributing

Fork it, fix it, push it, PR it.
