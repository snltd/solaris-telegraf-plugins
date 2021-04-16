package solaris_memory

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sh "github.com/snltd/sunos_telegraf_helpers"
	"log"
	"regexp"
	"strconv"
)

var sampleConfig = `
  ## General fields.
  # Fields = ["kernel", "arcsize", "freelist"]
	## Fields you want from from the 'unix:0:vminfo' kstats. These
	## should be turned into a rate by your graphing software
	# VmInfoFields = ["freemem", "swap_resv", "swap_alloc",
	#                 "swap_avail", "swap_free"]
	#
	## Fields from the output of 'swap -l'
	# SwapFields = ["allocated", "reserved", "used", "available"]
	## which swap-related fields you want from the cpu::vm kstat.
	## These should be turned into rates by your graphing software
	# CpuVmFields = []
	## Whether to aggregate CpuVmFields, or keep them separate
	# PerCpuVm = true
`

func (s *SolarisMemory) Description() string {
	return "Reports on Solaris virtual and physical memory usage"
}

func (s *SolarisMemory) SampleConfig() string {
	return sampleConfig
}

type SolarisMemory struct {
	Fields       []string
	VmInfoFields []string
	SwapFields   []string
	CpuVmFields  []string
	PerCpuVm     bool
}

func pageSize() int {
	pgsize, _ := strconv.Atoi(sh.RunCmd("/bin/pagesize"))
	return pgsize
}

func (s *SolarisMemory) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	pgsize := uint64(pageSize())
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	// miscellaneous memory stats

	if sh.WeWant("kernel", s.Fields) {
		kpg := sh.KstatSingle(token, "unix:0:system_pages:pp_kernel")
		fields["kernel"] = kpg.(uint64) * pgsize
	}

	if sh.WeWant("arcsize", s.Fields) {
		fields["arcsize"] = sh.KstatSingle(token, "zfs:0:arcstats:size")
	}

	if sh.WeWant("freelist", s.Fields) {
		pfree := sh.KstatSingle(token, "unix:0:system_pages:pagesfree")
		fields["freelist"] = pfree.(uint64) * pgsize
	}

	// vminfo kstats

	if len(s.VmInfoFields) > 0 {
		_, vi, err := token.Vminfo()

		if err != nil {
			log.Fatal("cannot get vminfo kstats")
		}

		if sh.WeWant("freemem", s.VmInfoFields) {
			fields["vminfo.freemem"] = vi.Freemem * pgsize
		}

		if sh.WeWant("swap_alloc", s.VmInfoFields) {
			fields["vminfo.swap_alloc"] = vi.Alloc * pgsize
		}

		if sh.WeWant("swap_avail", s.VmInfoFields) {
			fields["vminfo.swap_avail"] = vi.Avail * pgsize
		}

		if sh.WeWant("swap_free", s.VmInfoFields) {
			fields["vminfo.swap_free"] = vi.Free * pgsize
		}

		if sh.WeWant("swap_resv", s.VmInfoFields) {
			fields["vminfo.swap_resv"] = vi.Resv * pgsize
		}
	}

	// Swap -s  stuff

	if len(s.SwapFields) > 0 {
		swapline := sh.RunCmd("/usr/sbin/swap -s")

		re := regexp.MustCompile(
			`total: (\d+)k [\w ]* \+ (\d+)k.*= (\d+)k used, (\d+)k.*$`)

		m := re.FindAllStringSubmatch(swapline, -1)[0]

		if sh.WeWant("allocated", s.SwapFields) {
			fields["swap.allocated"], _ = strconv.Atoi(m[1])
		}

		if sh.WeWant("reserved", s.SwapFields) {
			fields["swap.reserved"], _ = strconv.Atoi(m[2])
		}

		if sh.WeWant("used", s.SwapFields) {
			fields["swap.used"], _ = strconv.Atoi(m[3])
		}

		if sh.WeWant("available", s.SwapFields) {
			fields["swap.available"], _ = strconv.Atoi(m[4])
		}
	}

	// Swapping and paging stats

	cpu_stats := sh.KstatModule(token, "cpu")
	sums := make(map[string]uint64)

	for _, name := range cpu_stats {
		if name.Name != "vm" {
			continue
		}

		stats, _ := name.AllNamed()

		for _, stat := range stats {
			if !sh.WeWant(stat.Name, s.CpuVmFields) {
				continue
			}

			if s.PerCpuVm {
				fkey := fmt.Sprintf("cpu.vm.%d.%s", stat.KStat.Instance, stat.Name)
				fields[fkey] = stat.UintVal
			} else {
				sums[stat.Name] = sums[stat.Name] + stat.UintVal
			}

		}

		if !s.PerCpuVm {
			for k, v := range sums {
				if sh.WeWant(k, s.CpuVmFields) {
					fkey := fmt.Sprintf("vm.%s", k)
					fields[fkey] = v
				}
			}
		}

	}

	acc.AddFields("solaris_memory", fields, tags)
	token.Close()
	return nil
}

func init() {
	inputs.Add("solaris_memory", func() telegraf.Input {
		return &SolarisMemory{}
	})
}
