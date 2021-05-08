package illumos_disk_health

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCamelCase(t *testing.T) {
	assert.Equal(t, "softErrors", camelCase("Soft Errors"))
	assert.Equal(t, "softErrors", camelCase("soft Errors"))
	assert.Equal(t, "word", camelCase("word"))
	assert.Equal(t, "oneTwoThree", camelCase("One tWo three"))
}

func TestParseNamedStats(t *testing.T) {
	s := &IllumosDiskHealth{
		Devices: []string{"sd6"},
		Fields:  []string{"Hard Errors", "Soft Errors", "Transport Errors", "Illegal Request"},
		Tags:    []string{"Vendor", "Serial No", "Product", "Revision"},
	}

	testData := sth.FromFixture("sderr:6:sd6,err.kstat")
	fields, tags := parseNamedStats(s, testData)

	assert.Equal(
		t,
		fields,
		map[string]interface{}{
			"hardErrors":      float64(0),
			"illegalRequest":  float64(1148),
			"softErrors":      float64(0),
			"transportErrors": float64(0),
		},
	)

	assert.Equal(
		t,
		tags,
		map[string]string{
			"product":  "My Passport 2627",
			"revision": "4008",
			"serialNo": "WXP1E7916Z6K",
			"vendor":   "WD",
		},
	)
}

func TestPlugin(t *testing.T) {
	s := &IllumosDiskHealth{
		Devices: []string{"sd6"},
		Fields:  []string{"Hard Errors", "Transport Errors", "Illegal Request"},
		Tags:    []string{"Vendor", "Serial No", "Product"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetrics,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime(),
	)
}

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"diskHealth",
		map[string]string{
			"product":  "My Passport 2627",
			"serialNo": "WXP1E7916Z6K",
			"vendor":   "WD",
		},
		map[string]interface{}{
			"hardErrors":      float64(0),
			"transportErrors": float64(0),
			"illegalRequest":  float64(1203),
		},
		time.Now(),
	),
}
