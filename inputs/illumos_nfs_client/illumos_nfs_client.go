package illumos_nfs_client

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sh "github.com/snltd/solaris-telegraf-helpers"
	"log"
	"strings"
)

var sampleConfig = `
	## The NFS versions you wish to monitor
	# nfs_versions = ["v3", "v4"]
  ## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfs' lists the possibilities
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

	ks := sh.KstatModule(token, "nfs")

	for _, stat := range ks {
		if !strings.HasPrefix(stat.Name, "rfsreqcnt_v") {
			continue
		}

		nfsVersion := fmt.Sprintf("v%s", stat.Name[len(stat.Name)-1:])

		if !sh.WeWant(nfsVersion, s.NfsVersions) {
			continue
		}

		stats, err := stat.AllNamed()

		if err != nil {
			log.Fatal("cannot get named NFS kstats")
		}

		fields := make(map[string]interface{})

		for _, stat := range stats {
			if !sh.WeWant(stat.Name, s.Fields) {
				continue
			}

			// cannot type switch on non-interface value stat (type *kstat.Named), hence this hack
			valueType := fmt.Sprintf("%s", stat.Type)

			if strings.HasPrefix(valueType, "uint") {
				fields[stat.Name] = stat.UintVal
			} else if strings.HasPrefix(valueType, "int") {
				fields[stat.Name] = stat.IntVal
			}
		}

		acc.AddFields("nfs.client", fields, map[string]string{"nfsVersion": nfsVersion})
	}

	token.Close()
	return nil
}

func init() {
	inputs.Add("illumos_nfs_client", func() telegraf.Input { return &IllumosNfsClient{} })
}
