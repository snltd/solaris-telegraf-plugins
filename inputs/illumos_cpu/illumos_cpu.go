package illumos_cpu

/*
Collects information about Illumos CPU usage. The values it outputs are the raw kstat values,
which means they are counters, and they only go up. I wrap them in a rate() function in
Wavefront, which is plenty good enough for me.
*/

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"log"
	"strings"
)

var sampleConfig = `
  ## report stuff from the cpu_info kstat. As of now it's just the current clock speed and some
	## potentially useful tags
	# cpu_info_stats = true
	## Produce metrics for sys and user CPU consumption in every zone
	# cpu_zone_stats = true
	## Which cpu:sys kstat metrics you wish to emit. They probably won't all work, because they
	## some will have a value type which is not an unsigned int
	# fields = ["cpu_nsec_dtrace", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"]
	## "cpu_ticks_idle", cpu_ticks_kernel", cpu_ticks_user", cpu_ticks_wait", }
`

func (s *IllumosCpu) Description() string {
	return "Reports on Illumos CPU usage"
}

func (s *IllumosCpu) SampleConfig() string {
	return sampleConfig
}

type IllumosCpu struct {
	ZoneCpuStats bool
	CpuInfoStats bool
	Fields       []string
}

var runPsrinfoCmd = func() string {
	return sth.RunCmd("/usr/sbin/psrinfo")
}

// cpuinfoTags extracts useful tags from a cpu_info kstat
func cpuinfoTags(stat *kstat.KStat) map[string]string {
	coreID, _ := stat.GetNamed("core_id")
	chipID, _ := stat.GetNamed("chip_id")
	state, _ := stat.GetNamed("state")
	clockMhz, _ := stat.GetNamed("clock_MHz")

	return map[string]string{
		"coreID":   fmt.Sprintf("%d", coreID.IntVal),
		"chipID":   fmt.Sprintf("%d", chipID.IntVal),
		"state":    state.StringVal,
		"clockMhz": fmt.Sprintf("%d", clockMhz.IntVal),
	}
}

func gatherCpuinfoStats(acc telegraf.Accumulator, token *kstat.Token) {
	stats := sth.KStatsInModule(token, "cpu_info")

	for _, stat := range stats {
		currentHz, err := stat.GetNamed("current_clock_Hz")

		if err == nil {
			acc.AddFields(
				"cpu",
				map[string]interface{}{"speed": currentHz.UintVal},
				cpuinfoTags(stat),
			)
		}
	}
}

func gatherCpuStats(s *IllumosCpu, acc telegraf.Accumulator, token *kstat.Token) {
	cpuStats := sth.KStatsInModule(token, "cpu")

	for _, statGroup := range cpuStats {
		if statGroup.Name != "sys" {
			continue
		}

		fields := make(map[string]interface{})

		for _, field := range s.Fields {
			stat, err := statGroup.GetNamed(field)

			if err != nil {
				log.Printf(fmt.Sprintf("cannot get %s:%s", statGroup, field))
			}

			fields[fieldToMetricPath(field)] = stat.UintVal
		}

		tags := map[string]string{
			"coreID": fmt.Sprintf("%d", statGroup.Instance),
		}

		acc.AddFields("cpu", fields, tags)
	}
}

func fieldToMetricPath(field string) string {
	field = strings.Replace(field, "cpu_", "", 1)
	field = strings.Replace(field, "_", ".", 1)
	return field
}

// metrics reporting on CPU consumption for each zone. sys and user, each as a gauge, tagged with
// the zone name
func gatherZoneCpuStats(acc telegraf.Accumulator, token *kstat.Token) {
	zoneStats := sth.KstatModule(token, "zones")

	for _, zone := range zoneStats {
		nsecSys, serr := zone.GetNamed("nsec_sys")
		nsecUser, uerr := zone.GetNamed("nsec_user")

		fields := map[string]interface{}{
			"sys":  nsecSys.UintVal,
			"user": nsecUser.UintVal,
		}

		tags := map[string]string{"zone": zone.Name}

		if serr == nil && uerr == nil {
			acc.AddFields("cpu.zone", fields, tags)
		}
	}
}

func (s *IllumosCpu) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	if s.CpuInfoStats {
		gatherCpuinfoStats(acc, token)
	}

	gatherCpuStats(s, acc, token)

	if s.ZoneCpuStats {
		gatherZoneCpuStats(acc, token)
	}

	token.Close()
	return nil
}

func init() {
	inputs.Add("illumos_cpu", func() telegraf.Input { return &IllumosCpu{} })
}
