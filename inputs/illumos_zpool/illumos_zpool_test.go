package illumos_zpool

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func TestHealthtoi(t *testing.T) {
	assert.Equal(t, 0, healthtoi("ONLINE"))
	assert.Equal(t, 1, healthtoi("DEGRADED"))
	assert.Equal(t, 2, healthtoi("SUSPENDED"))
	assert.Equal(t, 3, healthtoi("UNAVAIL"))
	assert.Equal(t, 99, healthtoi("what the heck is this nonsense"))
}

func TestParseZpool(t *testing.T) {
	line := "big    3.62T  2.69T   959G        -         -     2%    74%  1.00x  ONLINE  -"

	assert.Equal(
		t,
		Zpool{
			name: "big",
			props: map[string]interface{}{
				"size":   3.98023209254912e+12,
				"alloc":  2.95768627871744e+12,
				"free":   1.029718409216e+12,
				"frag":   2,
				"cap":    74,
				"dedup":  1.0,
				"health": 0,
			},
		},
		parseZpool(line, header),
	)
}

func TestParseHeader(t *testing.T) {
	assert.Equal(
		t,
		[]string{"name", "size", "alloc", "free", "ckpoint", "expandsz", "frag", "cap", "dedup",
			"health", "altroot"},
		parseHeader(header))
}

func TestPluginAllMetrics(t *testing.T) {
	s := &IllumosZpool{}

	zpoolOutput = func() string {
		return sampleOutput
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetricsFull,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetricsFull = []telegraf.Metric{
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "big",
		},
		map[string]interface{}{
			"size":   3.98023209254912e+12,
			"alloc":  2.95768627871744e+12,
			"free":   1.029718409216e+12,
			"frag":   2,
			"cap":    74,
			"dedup":  1.0,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "fast",
		},
		map[string]interface{}{
			"size":   2.81320357888e+11,
			"alloc":  1.11669149696e+11,
			"free":   1.69651208192e+11,
			"frag":   25,
			"cap":    39,
			"dedup":  1.0,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "rpool",
		},
		map[string]interface{}{
			"size":   2.13674622976e+11,
			"alloc":  6.13106581504e+10,
			"free":   1.52471339008e+11,
			"frag":   63,
			"cap":    28,
			"dedup":  1.0,
			"health": 0,
		},
		time.Now(),
	),
}

func TestPluginSelectedMetrics(t *testing.T) {
	s := &IllumosZpool{
		Fields: []string{"cap", "health"},
	}

	zpoolOutput = func() string {
		return sampleOutput
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetricsSelected,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetricsSelected = []telegraf.Metric{
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "big",
		},
		map[string]interface{}{
			"cap":    74,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "fast",
		},
		map[string]interface{}{
			"cap":    39,
			"health": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zpool",
		map[string]string{
			"name": "rpool",
		},
		map[string]interface{}{
			"cap":    28,
			"health": 0,
		},
		time.Now(),
	),
}

var header = "NAME    SIZE  ALLOC   FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP  HEALTH  ALTROOT"

var sampleOutput = `NAME    SIZE  ALLOC   FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP  HEALTH  ALTROOT
big    3.62T  2.69T   959G        -         -     2%    74%  1.00x  ONLINE  -
fast    262G   104G   158G        -         -    25%    39%  1.00x  ONLINE  -
rpool   199G  57.1G   142G        -         -    63%    28%  1.00x  ONLINE  -`
