# Illumos Net Input Plugin

Gathers high-level metrics about network traffic through Illumos NICs and
VNICs.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
[[illumos_network]]
  ## The kstat fields you wish to emit. 'kstat -c net' will show what is collected.  Not defining
  ## any fields sends everything, which is probably not what you want.
  # fields = ["obytes64", "rbytes64"]
  ## The VNICs you wish to observe. Again, specifying none collects all.
  # vnics  = ["net0"]
  ## The zones you wish to monitor. Specifying none collects all.
  # zones = ["zone1", "zone2"]`
```

Omitting `Fields` entirely results in all metrics being sent.

### Metrics

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

```
rate(ts("dev.telegraf.net.rbytes64", zone="global")) # bytes into your global zone
```

### Example Output

```
> net,host=cube,link=rge0,name=build_net0,speed=1000mbit,zone=cube-build rbytes64=1594089398i 1619391415000000000
> net,host=cube,link=rge0,name=build_net0,speed=1000mbit,zone=cube-build obytes64=208580451i 1619391415000000000
```
