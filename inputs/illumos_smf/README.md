# Illumos SMF Input Plugin

Gathers high-level metrics about SMF services on an Illumos system.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
[[inputs.illumos_smf]]
  ## The service states you wish to count.
  # svc_states = ["online", "uninitialized", "degraded", "maintenance"]
  ## The Zones you wish to examine. If this is unset or empty, all visible zones are counted.
  # zones = ["zone1", "zone2"]
  ## Whether or not you wish to generate individual, detailed points for services which are in
  ## SvcStates but are not "online"
  # generate_details = true
```

If it is running in the global zone, this plugin is able to collect SMF
information for all NGZs. However, the user running Telegraf must have the
`file_dac_search` privilege. `pfexec(1)` is used to gather information.

This plugin does not work on Solaris.

### Metrics

- smf
  - fields:
    - states (int, count of services defined by the following tags)
  - tags:
    - zone (string, zone name)
    - state (string, SMF service state)

  - fields:
    - errors (int, always 1)
  - tags:
    - fmri (string, service FMRI)
    - state (string, state the service is in)
    - zone (zone to which service belongs)


### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

To track all services which are not online:

```
ts("dev.telegraf.smf.states", NOT state="online")
```

To get detailed information on errant services. (Assuming `generate_details`
is true.)

```
ts("dev.telegraf.smf.errors")
```


### Example Output

```
> smf,host=cube,state=online,zone=cube-media states=61i 1619280366000000000
> smf,host=cube,state=online,zone=cube-pkgsrc states=45i 1619280366000000000
> smf,host=cube,state=uninitialized,zone=cube-pkgsrc states=3i 1619280366000000000
> smf,host=cube,state=maintenance,zone=cube-pkgsrc states=1i 1619280366000000000
> smf,host=cube,state=online,zone=cube-www-cassingle states=60i 1619280366000000000
> smf,host=cube,state=online,zone=cube-mariadb states=61i 1619280366000000000
> smf,host=cube,state=online,zone=cube-www-records states=60i 1619280366000000000
> smf,host=cube,state=online,zone=cube-dns states=58i 1619280366000000000
> smf,host=cube,state=online,zone=cube-pkg states=61i 1619280366000000000
> smf,host=cube,state=online,zone=cube-build states=62i 1619280366000000000
> smf,host=cube,state=online,zone=cube-www-proxy states=60i 1619280366000000000
> smf,host=cube,state=online,zone=cube-puppet states=57i 1619280366000000000
> smf,fmri=svc:/network/security/ktkt_warn:default,host=cube,state=uninitialized,zone=cube-pkgsrc errors=1i 1619280366000000000
> smf,fmri=svc:/network/rpc/gss:default,host=cube,state=uninitialized,zone=cube-pkgsrc errors=1i 1619280366000000000
> smf,fmri=svc:/network/nfs/rquota:default,host=cube,state=uninitialized,zone=cube-pkgsrc errors=1i 1619280366000000000
> smf,fmri=svc:/system/filesystem/local:default,host=cube,state=maintenance,zone=cube-pkgsrc errors=1i 1619280366000000000

```
