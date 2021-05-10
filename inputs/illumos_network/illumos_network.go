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

func (s *IllumosNetwork) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	links := sth.KStatsInModule(token, "link")

	for _, link := range links {
		// links are of the form link:0:dns_net0 for non-global zones, and link:0:rge0 (net) for the
		// global. (On Solaris the module number corresponds to the zone ID, but not on Illumos.)
		stats, _ := link.AllNamed()

		if err != nil {
			log.Fatal("cannot get named link kstats")
		}

		vnicMap := makeZoneVnicMap()
		vnic := vnicMap[link.Name]
		zone := vnic.Zone

		// If our vnicMap can't tell us which zone this belongs to, let's assume that it belongs to
		// the current zone. This might need to be smarter, but it's a reasonable first step. It
		// might be nice to pull some info about the physical NIC out into tags.
		if zone == "" {
			zone = zoneName
		}

		if !sth.WeWant(zone, s.Zones) {
			continue
		}

		acc.AddFields(
			"net",
			parseNamedStats(s, stats),
			zoneTags(zone, link.Name, vnic),
		)
	}

	token.Close()
	return nil
}

func zoneTags(zone, link string, vnic sth.Vnic) map[string]string {
	if zone == zoneName {
		return map[string]string{
			"zone":  zoneName,
			"link":  "none",
			"speed": "unknown",
			"name":  link,
		}
	}

	return map[string]string{
		"zone":  vnic.Zone,
		"link":  vnic.Link,
		"speed": fmt.Sprintf("%dmbit", vnic.Speed),
		"name":  vnic.Name,
	}
}

func parseNamedStats(s *IllumosNetwork, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {

		if !sth.WeWant(stat.Name, s.Fields) || !sth.WeWant(stat.KStat.Name, s.Vnics) { //||
			continue
		}

		fields[stat.Name] = sth.NamedValue(stat).(float64)
	}

	return fields
}

func init() {
	zoneName = sth.ZoneName()
	inputs.Add("illumos_network", func() telegraf.Input { return &IllumosNetwork{} })
}
