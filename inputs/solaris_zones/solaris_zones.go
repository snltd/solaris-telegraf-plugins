package solaris_zones

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/solaris-telegraf-helpers"
	"strings"
)

var sampleConfig = `
  ## The zone states you wish to report. If this is unset or an empty
	## array, all visible states will be collected. States which do not
	## occur in the output of zoneadm(1) will not be reported.
  ZoneStates = ["configured", "down", "incomplete", "installed", "ready",
                "running", "shutting_down"]
`

type SolarisZones struct {
	ZoneStates []string
}

func (z *SolarisZones) Description() string {
	return "Reports on Solaris zones."
}

func (z *SolarisZones) SampleConfig() string {
	return sampleConfig
}

func (z *SolarisZones) count_states(states string) map[string]int {
	ret := make(map[string]int)

	for _, s := range strings.Split(states, "\n") {
		ret[s] += 1
	}

	return ret
}

func (z *SolarisZones) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	states := make(map[string]int)
	brands := make(map[string]int)

	if len(z.ZoneStates) > 0 {
		for _, state := range z.ZoneStates {
			states[state] = 0
		}
	}

	for _, z := range sh.Zoneadm() {
		chunks := strings.Split(z, ":")
		states[chunks[2]] += 1
		brands[chunks[5]] += 1
	}

	for k, v := range states {
		fields[fmt.Sprintf("state.%s", k)] = v
	}

	for k, v := range brands {
		fields[fmt.Sprintf("brand.%s", k)] = v
	}

	acc.AddFields("zones", fields, nil)
	return nil
}

func init() {
	inputs.Add("zones", func() telegraf.Input { return &SolarisZones{} })
}
