# Illumos NFS Client Input Plugin

Gathers kstat metrics relating to an Illumos system's NFS client traffic. It
works with any NFS server version.

The kstat values are reported "raw": that is `crtime` and `snaptime` are not
used to calculate differentials. Your graphing software should calculate
rates, but they will not be as accurate as if they were calculated from the
high-resolution kstat times.

Telegraf minimum version: Telegraf 1.18
Plugin minimum tested version: 1.18

### Configuration

```toml
	## The NFS versions you wish to monitor.
	#NfsVersions = ["v3", "v4"]
	## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfs' lists the possibilities
	#Fields = ["read", "write", "remove", "create", "getattr", "setattr"]
```

Omitting `Fields` entirely results in all metrics being sent.

### Metrics

- zpool
  - tags:
    - nfsVersion (NFS protocol major version, e.g. "v4")
  - fields:
    - read (uint64)
    - ...

The final field of any `nfs:0:rfs*` kstat is a valid field.

### Sample Queries

The following queries are written in [The Wavefront Query
Language](https://docs.wavefront.com/query_language_reference.html).

```
rate(ts("dev.telegraf.nfs.client.write", nfsVersion="v4")) # write ops for NFSv4
rate(ts("dev.telegraf.nfs.client.read")) # all reads
```

### Example Output

```
> nfs.client,host=cube,nfsVersion=v3 create=0i,getattr=122i,read=194816i,remove=0i,setattr=0i,write=0i 1618958834000000000
> nfs.client,host=cube,nfsVersion=v4 create=291i,getattr=34952i,read=10793i,remove=1930i,setattr=854i,write=987i 1618958834000000000
```
