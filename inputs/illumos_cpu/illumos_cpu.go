package illumos_cpu

/*
Collects information about Illumos CPU usage. The values it outputs are the raw kstat values,
which means they are counters, and they only go up. I wrap them in a rate() function in
Wavefront, which is plenty good enough for me.
*/

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sth "github.com/snltd/solaris-telegraf-helpers"
)

var sampleConfig = `
  ## report stuff from the cpu_info kstat. As of now it's just the current clock speed and some
	## potentially useful tags
	# cpu_info_stats = true
	## Produce metrics for sys and user CPU consumption in every zone
	# zone_cpu_stats = true
	## Which cpu:sys kstat metrics you wish to emit. They probably won't all work, because they
	## some will have a value type which is not an unsigned int
	# sys_fields = ["cpu_nsec_dtrace", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"]
	## "cpu_ticks_idle", cpu_ticks_kernel", cpu_ticks_user", cpu_ticks_wait", }
`

func (s *IllumosCPU) Description() string {
	return "Reports on Illumos CPU usage"
}

func (s *IllumosCPU) SampleConfig() string {
	return sampleConfig
}

type IllumosCPU struct {
	CPUInfoStats bool
	ZoneCPUStats bool
	SysFields    []string
}

func parseCPUinfoKStats(stats []*kstat.Named) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, stat := range stats {
		switch stat.Name {
		case "current_clock_Hz":
			fields["speed"] = float64(stat.UintVal)
		case "clock_MHz":
			tags["clockMHz"] = fmt.Sprintf("%d", stat.IntVal)
		case "state":
			tags["state"] = stat.StringVal
		case "chip_id":
			tags["chipID"] = fmt.Sprintf("%d", stat.IntVal)
		case "core_id":
			tags["coreID"] = fmt.Sprintf("%d", stat.IntVal)
		}
	}

	return fields, tags
}

func gatherCPUinfoStats(acc telegraf.Accumulator, token *kstat.Token) {
	stats := sth.KStatsInModule(token, "cpu_info")

	for _, stat := range stats {
		namedStats, err := stat.AllNamed()
		if err != nil {
			log.Fatal("cannot get kstat token")
		}

		fields, tags := parseCPUinfoKStats(namedStats)
		acc.AddFields("cpu.info", fields, tags)
	}
}

func parseZoneCPUKStats(stats []*kstat.Named) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, stat := range stats {
		switch stat.Name {
		case "nsec_sys":
			fields["sys"] = float64(stat.UintVal)
		case "nsec_user":
			fields["user"] = float64(stat.UintVal)
		case "zonename":
			tags["name"] = stat.StringVal
		}
	}

	return fields, tags
}

// metrics reporting on CPU consumption for each zone. sys and user, each as a gauge, tagged with
// the zone name.
func gatherZoneCPUStats(acc telegraf.Accumulator, token *kstat.Token) {
	zoneStats := sth.KStatsInModule(token, "zones")

	for _, zone := range zoneStats {
		namedStats, err := zone.AllNamed()
		if err != nil {
			log.Fatal("cannot get zone CPU named stats")
		}

		fields, tags := parseZoneCPUKStats(namedStats)

		acc.AddFields("cpu.zone", fields, tags)
	}
}

func parseSysCPUKStats(s *IllumosCPU, stats []*kstat.Named) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, stat := range stats {
		if sth.WeWant(stat.Name, s.SysFields) {
			fields[fieldToMetricPath(stat.Name)] = float64(stat.UintVal)
		}
	}

	return fields
}

func gatherSysCPUStats(s *IllumosCPU, acc telegraf.Accumulator, token *kstat.Token) {
	cpuStats := sth.KStatsInModule(token, "cpu")

	for _, cpu := range cpuStats {
		if cpu.Name == "sys" {
			namedStats, err := cpu.AllNamed()
			if err != nil {
				log.Fatal("cannot get CPU named stats")
			}

			acc.AddFields(
				"cpu",
				parseSysCPUKStats(s, namedStats),
				map[string]string{"coreID": fmt.Sprintf("%d", cpu.Instance)},
			)
		}
	}
}

func fieldToMetricPath(field string) string {
	field = strings.Replace(field, "cpu_", "", 1)
	field = strings.Replace(field, "_", ".", 1)

	return field
}

func (s *IllumosCPU) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	if s.CPUInfoStats {
		gatherCPUinfoStats(acc, token)
	}

	if s.ZoneCPUStats {
		gatherZoneCPUStats(acc, token)
	}

	gatherSysCPUStats(s, acc, token)
	token.Close()

	return nil
}

func init() {
	inputs.Add("illumos_cpu", func() telegraf.Input { return &IllumosCPU{} })
}
