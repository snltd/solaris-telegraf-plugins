package illumos_zones

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

func (z *IllumosZones) Description() string {
	return "Report on zone states, brands, and other properties."
}

var (
	sampleConfig = ""
	makeZoneMap  = sth.NewZoneMap
)

type IllumosZones struct{}

func (z *IllumosZones) Gather(acc telegraf.Accumulator) error {
	gatherProperties(acc, makeZoneMap())

	return nil
}

func running(state string) int {
	if state == "running" {
		return 1
	}

	return 0
}

// Create an "I am here" metric for each zone. Value is 1 if the zone is running, 0 if it's not.
func gatherProperties(acc telegraf.Accumulator, zonemap sth.ZoneMap) {
	for zone, zoneData := range zonemap {
		acc.AddFields(
			"zones",
			map[string]interface{}{"status": running(zoneData.Status)},
			map[string]string{
				"name":   zone,
				"status": zoneData.Status,
				"ipType": zoneData.IPType,
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
