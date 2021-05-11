package smartos_zone

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sh "github.com/snltd/solaris-telegraf-helpers"

	//"strconv"
	"strings"
)

var sampleConfig = `
  ## This plugin outputs raw kstats. Most things will need to be
	## wrapped in a rate() type function in your graphing software.
  ## Use 'kstat -pc zone_caps' to see what kstat names are available,
	## and select the ones you want. You get the 'usage' (what you are using
  ## at this moment) and 'value' (the maximum available to you) values.
	# Names = ["swapresv", "lockedmem", "nprocs", "cpucaps", "physicalmem"]
	## You just get "usage" and "value" fields for almost everything
	## in 'Names' , but ## cpucaps has more information. Select the
	## fields you want here.  There's no need to include 'usage' or
	## 'value', they'll be done anyway.
	# CpuCapsFields = ["above_base_sec", "above_sec", "baseline",
	# "below_sec", "burst_limit_sec", "bursting_sec", "effective",
	# "maxusage", "nwait"]
	## Fields you require from the 'memory_cap' kstat module. Use
	## 'kstat -pm memory_cap' to view them. The memory_cap module is not
	## available on Solaris. 'rss' and 'size' are gauges: they do not
	## need converting to rates
	# MemoryCapFields = ["anon_alloc_fail", "anonpgin", "crtime", "execpgin",
  # "fspgin", "n_pf_throttle", "n_pf_throttle_usec", "nover", "pagedout",
	# "pgpgin", "physcap", "rss", "swap", "swapcap"]
`

func (s *SmartOsZone) Description() string {
	return "Reports metrics particular to a SmartOS zone"
}

func (s *SmartOsZone) SampleConfig() string {
	return sampleConfig
}

/*
func zoneId() int {
	raw := sh.RunCmd("/usr/sbin/zoneadm list -p")
	id, _ := strconv.Atoi(strings.Split(raw, ":")[0])
	return id
}
*/

type SmartOsZone struct {
	Names           []string
	CpuCapsFields   []string
	MemoryCapFields []string
}

func (s *SmartOsZone) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	zone_caps := sh.KstatClass(token, "zone_caps")

	for _, name := range zone_caps {
		stats, _ := name.AllNamed()
		nice_name := strings.Split(name.Name, "_")[0]

		if !sh.WeWant(nice_name, s.Names) {
			continue
		}

		for _, stat := range stats {
			if stat.Name == "zonename" {
				tags["zone"] = stat.StringVal
				continue
			}

			field := fmt.Sprintf("%s.%s", nice_name, stat.Name)

			if stat.Name == "value" || stat.Name == "usage" {
				fields[field] = stat.UintVal
			}

			if nice_name == "cpucaps" && sh.WeWant(stat.Name, s.CpuCapsFields) {
				fields[field] = stat.UintVal
			}
		}
	}

	mem_caps := sh.KstatModule(token, "memory_cap")

	for _, name := range mem_caps {
		stats, _ := name.AllNamed()

		for _, stat := range stats {

			if !sh.WeWant(stat.Name, s.MemoryCapFields) {
				continue
			}

			field := fmt.Sprintf("memory_cap.%s", stat.Name)
			fields[field] = stat.UintVal
		}
	}

	acc.AddFields("smartos_zone", fields, tags)
	token.Close()
	return nil
}

func init() {
	inputs.Add("smartos_zone", func() telegraf.Input {
		return &SmartOsZone{}
	})
}
