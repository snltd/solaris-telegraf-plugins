package solaris_io

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sh "github.com/snltd/solaris-telegraf-helpers"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var sampleConfig = `
  ## The kstat fields you wish to emit. 'kstat -c disk' will show what
	## is collected.
	## Not defining any fields sends everything, which is probably not
	## what you want
	Fields = ["reads", "nread", "writes", "nwritten"]
	## Do not report on the following disks.
	# OmitDisks = ["zones"]
`

func (s *SolarisIO) Description() string {
	return "Reports on Solaris IO"
}

func (s *SolarisIO) SampleConfig() string {
	return sampleConfig
}

type SolarisIO struct {
	OmitDisks []string
	Fields    []string
}

func (s *SolarisIO) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	r := regexp.MustCompile("[0-9]+$")

	for name, stat := range sh.KstatIoClass(token, "disk") {

		if sh.WeWant(name, s.OmitDisks) {
			continue
		}

		v := reflect.ValueOf(*stat)

		for i := 0; i < v.NumField(); i++ {
			fields := make(map[string]interface{})

			metric := strings.ToLower(v.Type().Field(i).Name)
			fname := fmt.Sprintf("%s.%s", name, metric)

			if !sh.WeWant(metric, s.Fields) {
				continue
			}

			// Tagging

			tags := make(map[string]string)
			num, err := strconv.Atoi(r.FindString(name))

			if err == nil {
				product := sh.KstatString(token, fmt.Sprintf(
					"sderr:%d:%s,err:Product", num, name))
				ser_no := sh.KstatString(token, fmt.Sprintf(
					"sderr:%d:%s,err:Serial No", num, name))

				if ser_no != "" {
					tags["ser_no"] = ser_no
				}

				if product != "" {
					tags["product"] = product
				}
			}

			val, err := strconv.Atoi(fmt.Sprintf("%d", v.Field(i)))

			if err != nil {
				log.Fatal("error converting")
			}

			fields[fname] = val
			acc.AddFields("solaris_io", fields, tags)
		}
	}

	token.Close()
	return nil
}

func init() {
	inputs.Add("solaris_io", func() telegraf.Input { return &SolarisIO{} })
}
