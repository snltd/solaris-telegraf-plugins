package illumos_fma

import (
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/solaris-telegraf-helpers"
)

var sampleConfig = `
	## Whether to report fmstat(1m) metrics
	# fmstat = true
	## Which fmstat modules to report
	# fmstat_modules = []
	## Which fmstat fields to report
	# fmstat_fields = []
	## Whether to report fmadm(1m) metrics
	# fmadm = true
`

type IllumosFma struct {
	Fmstat        bool
	FmstatModules []string
	FmstatFields  []string
	Fmadm         bool
}

type Fmstat struct {
	module string
	props  map[string]float64
}

func (s *IllumosFma) Description() string {
	return `A vague, experimental collector for the Illumos fault management architecture. I'm not
	sure yet what it is worth recording, and how, so this is almost certainly subject to change`
}

func (s *IllumosFma) SampleConfig() string {
	return sampleConfig
}

var runFmstatCmd = func() string {
	return sh.RunCmdPfexec("/usr/sbin/fmstat")
}

var runFmadmFaultyCmd = func() string {
	return sh.RunCmdPfexec("/usr/sbin/fmadm faulty")
}

func gatherFmstat(s *IllumosFma, acc telegraf.Accumulator) {
	raw := strings.Split(strings.TrimSpace(runFmstatCmd()), "\n")
	header := parseFmstatHeader(raw[0])

	for _, statLine := range raw[1:] {
		fields := make(map[string]interface{})
		fmstats := parseFmstatLine(statLine, header)

		if !sh.WeWant(fmstats.module, s.FmstatModules) {
			continue
		}

		for stat, val := range fmstats.props {
			if sh.WeWant(stat, s.FmstatFields) {
				fields[stat] = val
			}
		}

		acc.AddFields("fma.fmstat", fields, map[string]string{"module": fmstats.module})
	}
}

func gatherFmadm(acc telegraf.Accumulator) {
	fields := make(map[string]interface{})
	fmadmCounts := make(map[string]int)

	for _, impact := range fmadmImpacts() {
		safeName := strings.ReplaceAll(impact, ".", "_")
		fmadmCounts[safeName]++
	}

	for stat, value := range fmadmCounts {
		fields[stat] = value
	}

	acc.AddFields("fma.fmadm", fields, map[string]string{})
}

func (s *IllumosFma) Gather(acc telegraf.Accumulator) error {
	if s.Fmstat {
		gatherFmstat(s, acc)
	}

	if s.Fmadm {
		gatherFmadm(acc)
	}

	return nil
}

func parseFmstatHeader(headerLine string) []string {
	return strings.Fields(strings.ReplaceAll(headerLine, "%", "pc_"))
}

func parseFmstatLine(fmstatLine string, header []string) Fmstat {
	fields := strings.Fields(fmstatLine)
	props := make(map[string]float64)

	for i, field := range fields {
		property := header[i]

		switch property {
		case "module":
		case "memsz":
			fallthrough
		case "bufsz":
			props[property], _ = sh.Bytify(field)
		default:
			props[property], _ = strconv.ParseFloat(field, 64)
		}
	}

	return Fmstat{
		module: fields[0],
		props:  props,
	}
}

func fmadmImpacts() []string {
	raw := runFmadmFaultyCmd()
	lines := strings.Split(raw, "\n")

	var ret []string

	for _, line := range lines {
		if strings.Contains(line, "Problem class") {
			ret = append(ret, strings.Split(line, " : ")[1])
		}
	}

	return ret
}

func init() {
	inputs.Add("illumos_fma", func() telegraf.Input { return &IllumosFma{} })
}
