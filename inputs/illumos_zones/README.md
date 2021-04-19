# Illumos Zones Input Plugin

Gathers high-level metrics about the zones on an Illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

This plugin requires no configuration.

```toml
[[inputs.example]]
```

### Metrics

- zones
  - tags:
    - name (the zone name)
    - status (the zone status: `running`, `installed` etc.)
    - brand (the zone brand: `ipkg`, `lx`, `pkgsrc` etc.)
    - ipType (the zone's IP type: `excl`, `shared`)
  - fields:
    - status (int, `1` if the zone is running, `0` if it is not)

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

```
count(ts("dev.telegraf.zones.status"), status) # count the running/installed/etc zones
```


### Example Output

```
zones,brand=pkgsrc,host=cube,ipType=excl,name=cube-dns,status=running status=1i 1618866586000000000
zones,brand=lipkg,host=cube,ipType=excl,name=cube-pkg,status=running status=1i 1618866586000000000
zones,brand=pkgsrc,host=cube,ipType=excl,name=cube-ws,status=running status=1i 1618866586000000000

```
