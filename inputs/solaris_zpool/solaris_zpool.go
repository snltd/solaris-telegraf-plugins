package solaris_zpool

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/sunos_telegraf_helpers"
	"strconv"
	"strings"
)

var sampleConfig = `
  ## The metrics you wish to report. They can be any of the headers
	## in the output of 'zpool list', and 'health'.
	# Fields = ["size", "alloc", "free", "cap", "dedup", "health"]
	## Whether to count the number of dataset objects in a pool. Can be
	## any combination of 'filesystem', 'snapshot', and 'volume'.
	## Collecting this information can take a very long time if you have
	## large numbers of datasets. Put 'none' or something in there to
	## not do counts
	## Count = ["none"]
`

type SolarisZpool struct {
	Fields []string
}

func (s *SolarisZpool) Description() string {
	return "Reports the health and status of ZFS pools"
}

func (s *SolarisZpool) SampleConfig() string {
	return sampleConfig
}

// convert the health of a zpool to an integer, so you can alert off
// it.
// 0 : ONLINE
// 1 : DEGRADED
// 2 : SUSPENDED
// 3 : UNAVAIL
// 4 : <cannot parse>
//
func healthtoi(health string) int {
	states := []string{"ONLINE", "DEGRADED", "SUSPENDED", "UNAVAIL"}

	for i, state := range states {
		if state == health {
			return i
		}
	}

	return 4
}

// This has to be kind of vague, because the fields are different on
// different OSses
//
type Zpool struct {
	name  string
	props map[string]interface{}
}

// Return a struct describing a Zpool
//
func zpoolObject(zpool_line string, header []string) Zpool {
	fields := strings.Fields(zpool_line)
	ret := Zpool{name: fields[0]}

	props := make(map[string]interface{})

	for i, field := range fields {
		property := header[i]

		switch property {
		case "size":
			fallthrough
		case "alloc":
			fallthrough
		case "free":
			props[property], _ = sh.Bytify(field)
		case "cap":
			props["cap"], _ = strconv.Atoi(strings.TrimSuffix(field, "%"))
		case "dedup":
			strval := strings.TrimSuffix(field, "x")
			props["dedup"], _ = strconv.ParseFloat(strval, 64)
		case "health":
			props["health"] = healthtoi(field)
		default:
			props[property] = field
		}
	}

	ret.props = props
	return ret
}

func (s *SolarisZpool) Gather(acc telegraf.Accumulator) error {
	lines := strings.Split(sh.RunCmd("/usr/sbin/zpool list"), "\n")
	header := strings.Fields(strings.ToLower(lines[0]))
	fields := make(map[string]interface{})

	for _, pool := range lines[1:] {
		stats := zpoolObject(pool, header)
		tags := map[string]string{"name": stats.name}

		for stat, val := range stats.props {
			if sh.WeWant(stat, s.Fields) {
				fields[stat] = val
			}

		}

		acc.AddFields("solaris_zpool", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("solaris_zpool", func() telegraf.Input { return &SolarisZpool{} })
}
