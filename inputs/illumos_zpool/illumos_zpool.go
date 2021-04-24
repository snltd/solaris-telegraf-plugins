package illumos_zpool

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/solaris-telegraf-helpers"
	"strconv"
	"strings"
)

var sampleConfig = `
	## The metrics you wish to report. They can be any of the headers in the output of 'zpool list',
	## and also a numeric interpretation of 'health'.
	# fields = ["size", "alloc", "free", "cap", "dedup", "health"]
`

type IllumosZpool struct {
	Fields []string
}

func (s *IllumosZpool) Description() string {
	return "Reports the health and status of ZFS pools."
}

func (s *IllumosZpool) SampleConfig() string {
	return sampleConfig
}

var zpoolOutput = func() string {
	return sh.RunCmd("/usr/sbin/zpool list")
}

func (s *IllumosZpool) Gather(acc telegraf.Accumulator) error {
	raw := zpoolOutput()
	lines := strings.Split(raw, "\n")
	fields := make(map[string]interface{})

	for _, pool := range lines[1:] {
		poolStats := parseZpool(pool, lines[0])
		tags := map[string]string{"name": poolStats.name}

		for stat, val := range poolStats.props {
			if sh.WeWant(stat, s.Fields) {
				fields[stat] = val
			}
		}

		acc.AddFields("zpool", fields, tags)
	}

	return nil
}

// parseHeader turns the first line of `zpool list`'s output into an array of lower-case strings
func parseHeader(raw string) []string {
	return strings.Fields(strings.ToLower(raw))
}

// healthtoi converts the health of a zpool to an integer, so you can alert off it.
// 0 : ONLINE
// 1 : DEGRADED
// 2 : SUSPENDED
// 3 : UNAVAIL
// 4 : FAULTED
// 99: <cannot parse>
func healthtoi(health string) int {
	states := []string{"ONLINE", "DEGRADED", "SUSPENDED", "UNAVAIL", "FAULTED"}

	for i, state := range states {
		if state == health {
			return i
		}
	}

	return 99
}

// Zpool stores all the Zpool properties in the `props` map, which is dynamically generated. This
// means it will work on Solaris as well as Illumos, and won't break if the output format of
// `zpool(1m)` changes.
type Zpool struct {
	name  string
	props map[string]interface{}
}

// parseZpool semi-intelligently parses a line of `zpool list` output, using that command's output
// header to pick out the fields we are interested in.
func parseZpool(raw, rawHeader string) Zpool {
	fields := parseHeader(rawHeader)
	chunks := strings.Fields(raw)
	pool := Zpool{
		name:  chunks[0],
		props: make(map[string]interface{}),
	}

	for i, field := range chunks {
		property := fields[i]

		switch property {
		case "size":
			fallthrough
		case "alloc":
			fallthrough
		case "free":
			pool.props[property], _ = sh.Bytify(field)
		case "frag":
			fallthrough
		case "cap":
			pool.props[property], _ = strconv.Atoi(strings.TrimSuffix(field, "%"))
		case "dedup":
			strval := strings.TrimSuffix(field, "x")
			pool.props["dedup"], _ = strconv.ParseFloat(strval, 64)
		case "health":
			pool.props["health"] = healthtoi(field)
		}
	}

	return pool
}

func init() {
	inputs.Add("illumos_zpool", func() telegraf.Input { return &IllumosZpool{} })
}
