package illumos_nfs_client

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

var sampleConfig = `
	## The NFS versions you wish to monitor
	# nfs_versions = ["v3", "v4"]
  ## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfsreqcnt' lists the
	## possibilities
	# fields = ["read", "write", "remove", "create", "getattr", "setattr"]
`

func (s *IllumosNfsClient) Description() string {
	return "Reports Illumos NFS client statistics"
}

func (s *IllumosNfsClient) SampleConfig() string {
	return sampleConfig
}

type IllumosNfsClient struct {
	Fields      []string
	NfsVersions []string
}

func (s *IllumosNfsClient) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	stats := sth.KStatsInModule(token, "nfs")

	for _, stat := range stats {
		if !strings.HasPrefix(stat.Name, "rfsreqcnt_v") {
			continue
		}

		nfsVersion := fmt.Sprintf("v%s", stat.Name[len(stat.Name)-1:])

		if !sth.WeWant(nfsVersion, s.NfsVersions) {
			continue
		}

		stats, err := stat.AllNamed()
		if err != nil {
			log.Fatal("cannot get named NFS client kstats")
		}

		acc.AddFields(
			"nfs.client",
			parseNamedStats(s, stats),
			map[string]string{"nfsVersion": nfsVersion},
		)
	}

	token.Close()

	return nil
}

func parseNamedStats(s *IllumosNfsClient, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if sth.WeWant(stat.Name, s.Fields) {
			fields[stat.Name] = sth.NamedValue(stat).(float64)
		}
	}

	return fields
}

func init() {
	inputs.Add("illumos_nfs_client", func() telegraf.Input { return &IllumosNfsClient{} })
}
