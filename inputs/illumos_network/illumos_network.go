package illumos_network

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"log"
)

var sampleConfig = `
	## The kstat fields you wish to emit. 'kstat -c net' will show what is collected.  Not defining
	## any fields sends everything, which is probably not what you want.
	# fields = ["obytes64", "rbytes64"]
	## The VNICs you wish to observe. Again, specifying none collects all.
	# vnics  = ["net0"]
	## The zones you wish to monitor. Specifying none collects all.
	# zones = ["zone1", "zone2"]`

func (s *IllumosNetwork) Description() string {
	return "Reports on Illumos NIC Usage. Zone-aware."
}

func (s *IllumosNetwork) SampleConfig() string {
	return sampleConfig
}

type IllumosNetwork struct {
	Zones  []string
	Fields []string
	Vnics  []string
}

var makeZoneVnicMap = func() sth.ZoneVnicMap {
	return sth.NewZoneVnicMap()
}

var zoneName = ""

var kstatData = func() (*kstat.Token, []*kstat.KStat) {
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	return token, sth.KstatModule(token, "link")
}

func (s *IllumosNetwork) Gather(acc telegraf.Accumulator) error {
	// To tag a VNIC with the zone that uses it, we need information from dladm(1m). I do this on
	// every run so we catch new zones coming or old ones going.
	vnicMap := makeZoneVnicMap()
	token, mods := kstatData()

	for _, mod := range mods {
		stats, _ := mod.AllNamed()

		for _, stat := range stats {
			// mods are of the form link:0:dns_net0 for non-global zones, and link:0:rge0 (net) for the
			// global. (On Solaris the module number corresponds to the zone ID, but not on Illumos.)
			vnic := vnicMap[stat.KStat.Name]
			zone := vnic.Zone

			// If our vnicMap can't tell us which zone this belongs to, let's assume that it belongs to
			// the current zone. This might need to be smarter, but it's a reasonable first step. It
			// might be nice to pull some info about the physical NIC out into tags.
			if zone == "" {
				zone = zoneName
			}

			if !sth.WeWant(stat.Name, s.Fields) || !sth.WeWant(stat.KStat.Name, s.Vnics) ||
				!sth.WeWant(zone, s.Zones) {
				continue
			}

			var tags map[string]string

			if zone == zoneName {
				tags = map[string]string{
					"zone":  zoneName,
					"link":  "none",
					"speed": "unknown",
					"name":  mod.Name,
				}
			} else {
				tags = map[string]string{
					"zone":  vnic.Zone,
					"link":  vnic.Link,
					"speed": fmt.Sprintf("%dmbit", vnic.Speed),
					"name":  vnic.Name,
				}
			}

			acc.AddFields(
				"net",
				map[string]interface{}{fmt.Sprintf("%s", stat.Name): stat.UintVal},
				tags,
			)
		}
	}

	token.Close()
	return nil
}

func init() {
	zoneName = sth.ZoneName()
	inputs.Add("illumos_network", func() telegraf.Input { return &IllumosNetwork{} })
}
