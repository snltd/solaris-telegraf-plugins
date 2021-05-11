package illumos_cpu

import (
	//"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	//"os"
	"testing"
	"time"
)

func TestFieldToMetricPath(t *testing.T) {
	assert.Equal(
		t,
		"nsec.kernel",
		fieldToMetricPath("cpu_nsec_kernel"),
	)
}

func _TestPlugin(t *testing.T) {
	s := &IllumosCpu{}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetrics,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"merp",
		map[string]string{},
		map[string]interface{}{},
		time.Now(),
	),
}

/*
func TestCountVcpus(t *testing.T) {
	runPsrinfoCmd = func() string {
		return samplePsrinfoOutput
	}

	assert.Equal(
		t,
		[]string{"0", "1", "2", "3"},
		countVcpus(),
	)
}

var samplePsrinfoOutput = `0	on-line   since 05/04/2021 11:07:12
1	on-line   since 05/04/2021 11:07:12
2	on-line   since 05/04/2021 11:07:13
3	on-line   since 05/04/2021 11:07:13`
*/
