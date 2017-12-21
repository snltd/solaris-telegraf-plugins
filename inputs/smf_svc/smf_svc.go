package smf_svc

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/sunos_helpers"
	"regexp"
	"strings"
)

var sampleConfig = `
  ## The service states you wish to report. If this is unset or an empty
	## array, all visible states will be collected. States which do not
	## occur in the output of svcs(1) will not be reported.
	# SvcStates = ["online", "uninitialized", "degraded", "maintenance"]
	#
	## Zones on which you wish to report. Omit or set empty to get all
	## native zones on the host. This only works on SmartOS, as Solaris's
	## svcs(1) command does not support the '-z' option.
	# Zones = ["zone1", "zone2"]
`

type SmfSvc struct {
	zone      string
	SvcStates []string
	Zones     []string
}

func (s *SmfSvc) Description() string {
	return "Reports the number of SMF services in each state"
}

func (s *SmfSvc) SampleConfig() string {
	return sampleConfig
}

func (s *SmfSvc) count_states(states string) map[string]int {
	ret := make(map[string]int)

	for _, s := range strings.Split(states, "\n") {
		ret[s] += 1
	}

	return ret
}

func (s *SmfSvc) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	// Zero all fields, otherwise you'll get no metric for, say
	// maintenance, if nothing's in maintenance

	if len(s.SvcStates) > 0 {
		for _, state := range s.SvcStates {
			fields[state] = 0
		}
	}

	// This plugin is zone-aware on SmartOS, but must only query
	// native zones.

	if regexp.MustCompile("-Z").MatchString(sh.RunCmd("/usr/bin/svcs -h")) {
		for _, z := range sh.Zoneadm() {

			props := strings.Split(z, ":")
			zone := props[1]

			if !sh.WeWant(zone, s.Zones) {
				continue
			}

			if props[5] != "joyent" {
				continue
			}

			raw := sh.RunCmd(fmt.Sprintf("/bin/svcs -z %s -aHo state", zone))
			tags["zone"] = zone

			for k, v := range s.count_states(raw) {
				if k != "" && sh.WeWant(k, s.SvcStates) {
					fields[k] = v
				}
			}

			acc.AddFields("smf_svc", fields, tags)
		}

	} else {
		// On Solaris we can only deal with the zone we're in, unless we
		// do some sort of awful pfexec/zlogin thing, and right now I'm
		// not doing that.
		var zone string

		// Cache the zone name. It can't change.
		if s.zone == "" {
			zone = sh.Zone()
			s.zone = zone
		} else {
			zone = s.zone
		}

		raw := sh.RunCmd("/bin/svcs -aHo state")
		tags["zone"] = zone

		for k, v := range s.count_states(raw) {
			if k != "" && sh.WeWant(k, s.SvcStates) {
				fields[k] = v
			}
		}

		acc.AddFields("smf_svc", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("smf_svc", func() telegraf.Input { return &SmfSvc{} })
}
