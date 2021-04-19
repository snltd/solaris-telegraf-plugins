package illumos_zones

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

func (z *IllumosZones) Description() string {
	return "Report on zone states, brands, and other properties"
}

var sampleConfig = `
	## Count the number of zones in the following states
	ZoneStates = ["configured", "down", "incomplete", "installed", "ready", "running",
	"shutting_down"]
	## Count the number of zones with the following brands
	ZoneBrands = ["ipkg", "lipkg", "pkgsrc", "bhyve", "kvm", "sparse", "sn1", "s10", "lx", "illumos"]
	# If ZoneProperties is true, an individual up/down metric is pr
	ZoneProperties = true
	ZoneCount = true
`

type IllumosZones struct {
	ZoneStates []string
	ZoneBrands []string
	ZoneID     bool
	ZoneCount  bool
}

var makeZoneMap = func() sth.ZoneMap {
	return sth.NewZoneMap()
}

func (z *IllumosZones) Gather(acc telegraf.Accumulator) error {
	acc.AddFields("zones", gatherCounts(z, makeZoneMap()), nil)
	if z.ZoneID {
		gatherProperties(z, acc, makeZoneMap())
	}
	return nil
}

// Count zones in requested states and of requested brand, and create a metric counting how many
// zones are visible, regardless of configuration.
func gatherCounts(z *IllumosZones, zonemap sth.ZoneMap) map[string]interface{} {
	fields := make(map[string]interface{})
	states := zeroedMap(z.ZoneStates)
	brands := zeroedMap(z.ZoneBrands)

	for _, zoneData := range zonemap {
		if _, ok := states[zoneData.Status]; ok {
			states[zoneData.Status] += 1
		}

		if _, ok := brands[zoneData.Brand]; ok {
			brands[zoneData.Brand] += 1
		}
	}

	for state, count := range states {
		fields[fmt.Sprintf("state.%s", state)] = count
	}

	for brand, count := range brands {
		fields[fmt.Sprintf("brand.%s", brand)] = count
	}

	if z.ZoneCount == true {
		fields["count"] = len(zonemap)
	}

	return fields
}

func running(state string) int {
	if state == "running" {
		return 1
	} else {
		return 0
	}
}

// Create an "I am here" metric for each zone. Value is 1 if the zone is running, 0 if it's not.
func gatherProperties(z *IllumosZones, acc telegraf.Accumulator, zonemap sth.ZoneMap) {
	for zone, zoneData := range zonemap {
		acc.AddFields(
			"zones",
			map[string]interface{}{"properties": running(zoneData.Status)},
			map[string]string{
				"name":   zone,
				"status": zoneData.Status,
				"ipType": zoneData.IpType,
				"brand":  zoneData.Brand,
			})
	}
}

func (z *IllumosZones) SampleConfig() string {
	return sampleConfig
}

// zeroedMap turns a list of strings into a map with the strings as the keys and 0 as the value
func zeroedMap(keys []string) map[string]int {
	ret := make(map[string]int, len(keys))

	for _, state := range keys {
		ret[state] = 0
	}

	return ret
}

func init() {
	inputs.Add("illumos_zones", func() telegraf.Input { return &IllumosZones{} })
}
