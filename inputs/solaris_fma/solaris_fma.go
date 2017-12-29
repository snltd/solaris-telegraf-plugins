package solaris_fma

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/sunos_helpers"
	"strconv"
	"strings"
)

var sampleConfig = `
	## Whether to report 'fmstat' metrics
	# Fmstat = true
	## Which 'fmstat' fields to report
	# FmstatFields = []
	## Whether to report 'fmadm' metrics
	# Fmadm = true
`

type SolarisFma struct {
	Fmstat       bool
	FmstatFields []string
	Fmadm        bool
}

func (s *SolarisFma) Description() string {
	return `A vague, experimental collector for the Solaris fault
	management architecture. I'm not sure yet what it is worth
	recording, and how, so this is almost certainly subject to change
	`
}

func (s *SolarisFma) SampleConfig() string {
	return sampleConfig
}

// return an array of faulty classes
//
func fmadmImpacts() []string {
	raw := strings.Split(sh.RunCmdPfexec("/usr/sbin/fmadm faulty"), "\n")
	var ret []string

	for _, line := range raw {
		if strings.Contains(line, "Problem class") {
			ret = append(ret, strings.Split(line, " : ")[1])
		}
	}

	return ret
}

// fmstat(1) output turned into a map where the module name is the
// key and the value is a struct of Fmstat
//
func fmstat() []Fmstat {
	raw := strings.Split(sh.RunCmdPfexec("/usr/sbin/fmstat"), "\n")
	header := fmStatHeader(raw[0])
	lines := raw[1:]

	var ret []Fmstat

	for _, line := range lines {
		ret = append(ret, fmStatObject(line, header))
	}

	return ret
}

type Fmstat struct {
	module string
	props  map[string]float64
}

// fmstat output header, as an array, with % signs changed to 'pc_'
//
func fmStatHeader(header_line string) []string {
	var ret []string

	for _, col := range strings.Fields(header_line) {
		if strings.HasPrefix(col, "%") {
			col = strings.Replace(col, "%", "pc_", -1)
		}
		ret = append(ret, col)
	}

	return ret
}

func fmStatObject(fma_line string, header []string) Fmstat {
	fields := strings.Fields(fma_line)
	ret := Fmstat{module: fields[0]}

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

	ret.props = props
	return ret
}

func (s *SolarisFma) Gather(acc telegraf.Accumulator) error {

	if s.Fmstat {
		raw := strings.Split(sh.RunCmdPfexec("/usr/sbin/fmstat"), "\n")
		header := fmStatHeader(raw[0])
		lines := raw[1:]

		for _, module := range lines[1:] {
			fields := make(map[string]interface{})
			stats := fmStatObject(module, header)
			tags := map[string]string{"name": stats.module}

			for stat, val := range stats.props {
				if sh.WeWant(stat, s.FmstatFields) {
					field := fmt.Sprintf("fmstat.%s", stat)
					fields[field] = val
				}

			}

			acc.AddFields("solaris_fma", fields, tags)
		}
	}

	if s.Fmadm {
		fields := make(map[string]interface{})
		var tags map[string]string

		fmadm_counts := make(map[string]int)

		for _, impact := range fmadmImpacts() {
			safe_name := strings.Replace(impact, ".", "_", -1)
			fmadm_counts[safe_name]++
		}

		for stat, value := range fmadm_counts {
			field := fmt.Sprintf("fmadm.%s", stat)
			fields[field] = value
		}

		acc.AddFields("solaris_fma", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("solaris_fma", func() telegraf.Input { return &SolarisFma{} })
}
