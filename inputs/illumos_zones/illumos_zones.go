package illumos_zones

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

func (z *IllumosZones) Description() string {
	return "Report on zone states, brands, and other properties."
}

var sampleConfig = ""

type IllumosZones struct{}

var makeZoneMap = func() sth.ZoneMap {
	return sth.NewZoneMap()
}

func (z *IllumosZones) Gather(acc telegraf.Accumulator) error {
	gatherProperties(z, acc, makeZoneMap())
	return nil
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
			map[string]interface{}{"status": running(zoneData.Status)},
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

func init() {
	inputs.Add("illumos_zones", func() telegraf.Input { return &IllumosZones{} })
}
