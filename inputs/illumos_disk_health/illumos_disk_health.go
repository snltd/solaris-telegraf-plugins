package illumos_disk_health

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
	## The kstat fields you wish to emit. 'kstat -c device_error' will show what is collected. Field
	## names will be camelCased in the metric path.
	# fields = ["Hard Errors", "Soft Errors", "Transport Errors", "Illegal Request"]
	## The tags you wish your data points to have. Not all devices are able to supply all tags, but
	## they will fail silently. Tag names are camelCased.
	# tags = ["Vendor", "Serial No", "Product", "Revision"]
	## Report on the following devices. Specifying none reports on all.
	# devices = ["sd6"]
`

func (s *IllumosDiskHealth) Description() string {
	return "Reports on Illumos disk errors"
}

func (s *IllumosDiskHealth) SampleConfig() string {
	return sampleConfig
}

type IllumosDiskHealth struct {
	Devices []string
	Fields  []string
	Tags    []string
}

// The info for the tags and the values is in the same kstat. There's no point going through it
// twice, so we'll return a tuple.
func parseNamedStats(s *IllumosDiskHealth, stats []*kstat.Named) (map[string]interface{}, map[string]string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, stat := range stats {
		switch {
		case sth.WeWant(stat.Name, s.Fields):
			fields[camelCase(stat.Name)] = sth.NamedValue(stat).(float64)
		case stat.Name == "Size" && sth.WeWant("Size", s.Tags):
			tags["size"] = fmt.Sprintf("%d", sth.NamedValue(stat))
		case sth.WeWant(stat.Name, s.Tags):
			tags[camelCase(stat.Name)] = strings.TrimSpace(stat.StringVal)
		}
	}

	return fields, tags
}

func (s *IllumosDiskHealth) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	statList := sth.KStatsInClass(token, "device_error")

	for _, stat := range statList {
		chunks := strings.Split(stat.Name, ",")
		deviceName := chunks[0]

		if sth.WeWant(deviceName, s.Devices) {
			namedStats, err := stat.AllNamed()

			if err == nil {
				fields, tags := parseNamedStats(s, namedStats)
				acc.AddFields("diskHealth", fields, tags)
			}
		}
	}

	token.Close()

	return nil
}

func camelCase(str string) string {
	words := strings.Fields(strings.Title(strings.ToLower(str)))
	words[0] = strings.ToLower(words[0])

	return strings.Join(words, "")
}

func init() {
	inputs.Add("illumos_disk_health", func() telegraf.Input { return &IllumosDiskHealth{} })
}
