package solaris_network

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sh "github.com/snltd/solaris-telegraf-helpers"
	"log"
)

var sampleConfig = `
  ## The kstat fields you wish to emit. 'kstat -c net' will show what
	## is collected.
	## Not defining any fields sends everything, which is probably not
	## what you want
	Fields = ["obytes64", "rbytes64"]
	## The NICs you wish to observe
	Nics  = ["net0"]
	## The zones you wish to monitor. Defaults to all
	# Zones = ["zone1", "zone2"]
`

func (s *SolarisNetwork) Description() string {
	return "Reports on Solaris NIC Usage"
}

func (s *SolarisNetwork) SampleConfig() string {
	return sampleConfig
}

type SolarisNetwork struct {
	Zones  []string
	Fields []string
	Nics   []string
}

func (s *SolarisNetwork) Gather(acc telegraf.Accumulator) error {
	zone_map := sh.ZoneMap()
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	mods := sh.KstatModule(token, "link")

	for _, mod := range mods {
		stats, _ := mod.AllNamed()

		for _, stat := range stats {
			if !sh.WeWant(stat.Name, s.Fields) {
				continue
			}

			if !sh.WeWant(stat.KStat.Name, s.Nics) {
				continue
			}

			zone := zone_map[stat.KStat.Instance]

			if !sh.WeWant(zone, s.Zones) {
				continue
			}

			fields := make(map[string]interface{})

			fname := fmt.Sprintf("%s", stat.Name)
			stat_type := fmt.Sprintf("%s", stat.Type)

			if stat_type == "uint32" || stat_type == "uint64" {
				fields[fname] = stat.UintVal
			} else if stat_type == "int32" || stat_type == "int64" {
				fields[fname] = stat.IntVal
			}

			tags := make(map[string]string)

			tags["zone"] = zone
			tags["nic"] = stat.KStat.Name

			acc.AddFields("solaris_network", fields, tags)
		}
	}

	token.Close()
	return nil
}

func init() {
	inputs.Add("solaris_network", func() telegraf.Input {
		return &SolarisNetwork{}
	})
}
