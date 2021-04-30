package illumos_memory

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"log"
	"regexp"
	"strconv"
)

var sampleConfig = `
  ## General fields.
  # fields = ["kernel", "arcsize", "freelist"]
	## Fields you want from from the 'unix:0:vminfo' kstats. These should be turned into a rate by
	## your graphing software
	# vm_info_fields = ["freemem", "swap_resv", "swap_alloc", "swap_avail", "swap_free"]
	## Fields from the output of 'swap -l'
	# swap_fields = ["allocated", "reserved", "used", "available"]
	## which swap-related fields you want from the 'cpu::vm' kstat. These should be turned into
	## rates by your graphing software
	# cpu_vm_fields = ["pgin", "anonpgin", "pgpgin", "pgout", "anonpgout", "pgpgout"]
	## If per_cpu_vm is true, you get all of the cpu_vm_fields for every VCPU. If it is false, they
	## are aggregated
	# per_cpu_vm = false
`

func (s *IllumosMemory) Description() string {
	return "Reports on Illumos virtual and physical memory usage."
}

func (s *IllumosMemory) SampleConfig() string {
	return sampleConfig
}

type IllumosMemory struct {
	Fields       []string
	VmInfoFields []string
	SwapFields   []string
	CpuVmFields  []string
	PerCpuVm     bool
}

var pageSize int

var runSwapCmd = func() string {
	return sth.RunCmd("/usr/sbin/swap -s")
}

var runPagesizeCmd = func() string {
	return sth.RunCmd("/bin/pagesize")
}

func systemPageSize() int {
	pageSize, err := strconv.Atoi(runPagesizeCmd())

	if err != nil {
		log.Fatal("Cannot get page size")
	}

	return pageSize
}

func (s *IllumosMemory) Gather(acc telegraf.Accumulator) error {
	tags := make(map[string]string)
	token, err := kstat.Open()

	if err != nil {
		return err
	}

	acc.AddFields("memory", miscKstats(s, token), tags)
	acc.AddFields("memory", swapAndPageKstats(s, token), tags)
	acc.AddFields("memory.vminfo", vminfoKstats(s, token), tags)

	if len(s.SwapFields) > 0 {
		acc.AddFields("memory.swap", parseSwap(s), tags)
	}

	token.Close()
	return nil
}

func miscKstats(s *IllumosMemory, token *kstat.Token) map[string]interface{} {
	fields := make(map[string]interface{})

	if sth.WeWant("kernel", s.Fields) {
		kpg := sth.KstatSingle(token, "unix:0:system_pages:pp_kernel")
		fields["kernel"] = float64(kpg.(uint64)) * float64(pageSize)
	}

	if sth.WeWant("arcsize", s.Fields) {
		fields["arcsize"] = sth.KstatSingle(token, "zfs:0:arcstats:size")
	}

	if sth.WeWant("freelist", s.Fields) {
		pfree := sth.KstatSingle(token, "unix:0:system_pages:pagesfree")
		fields["freelist"] = float64(pfree.(uint64)) * float64(pageSize)
	}

	return fields
}

// The raw kstats in here are gauges, measured in pages. So we need to convert them to bytes here,
// and you need to apply some kind of rate() function in your graphing software.
func vminfoKstats(s *IllumosMemory, token *kstat.Token) map[string]interface{} {
	fields := make(map[string]interface{})

	if len(s.VmInfoFields) > 0 {
		_, vi, err := token.Vminfo()

		if err != nil {
			log.Fatal("cannot get vminfo kstats")
		}

		if sth.WeWant("freemem", s.VmInfoFields) {
			fields["freemem"] = float64(vi.Freemem) * float64(pageSize)
		}

		if sth.WeWant("swap_alloc", s.VmInfoFields) {
			fields["swap_alloc"] = float64(vi.Alloc) * float64(pageSize)
		}

		if sth.WeWant("swap_avail", s.VmInfoFields) {
			fields["swap_avail"] = float64(vi.Avail) * float64(pageSize)
		}

		if sth.WeWant("swap_free", s.VmInfoFields) {
			fields["swap_free"] = float64(vi.Free) * float64(pageSize)
		}

		if sth.WeWant("swap_resv", s.VmInfoFields) {
			fields["swap_resv"] = float64(vi.Resv) * float64(pageSize)
		}
	}

	return fields
}

func swapAndPageKstats(s *IllumosMemory, token *kstat.Token) map[string]interface{} {
	fields := make(map[string]interface{})
	sums := make(map[string]uint64)
	cpu_stats := sth.KstatModule(token, "cpu")

	for _, name := range cpu_stats {
		if name.Name != "vm" {
			continue
		}

		stats, _ := name.AllNamed()

		for _, stat := range stats {
			if !sth.WeWant(stat.Name, s.CpuVmFields) {
				continue
			}

			if s.PerCpuVm {
				key := fmt.Sprintf("cpu.vm.%d.%s", stat.KStat.Instance, stat.Name)
				fields[key] = stat.UintVal
			} else {
				sums[stat.Name] = sums[stat.Name] + stat.UintVal
			}

		}

		if !s.PerCpuVm {
			for k, v := range sums {
				if sth.WeWant(k, s.CpuVmFields) {
					fkey := fmt.Sprintf("vm.%s", k)
					fields[fkey] = v
				}
			}
		}

	}

	return fields
}

func parseSwap(s *IllumosMemory) map[string]interface{} {
	fields := make(map[string]interface{})
	swapline := runSwapCmd()
	re := regexp.MustCompile(`total: (\d+k) [\w ]* \+ (\d+k).*= (\d+k) used, (\d+k).*$`)
	m := re.FindAllStringSubmatch(swapline, -1)[0]

	if sth.WeWant("allocated", s.SwapFields) {
		bytes, _ := sth.Bytify(m[1])
		fields["allocated"] = bytes
	}

	if sth.WeWant("reserved", s.SwapFields) {
		bytes, _ := sth.Bytify(m[2])
		fields["reserved"] = bytes
	}

	if sth.WeWant("used", s.SwapFields) {
		bytes, _ := sth.Bytify(m[3])
		fields["used"] = bytes
	}

	if sth.WeWant("available", s.SwapFields) {
		bytes, _ := sth.Bytify(m[4])
		fields["available"] = bytes
	}

	return fields
}

func init() {
	pageSize = systemPageSize()
	inputs.Add("illumos_memory", func() telegraf.Input { return &IllumosMemory{} })
}
