package illumos_cpu

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"github.com/stretchr/testify/require"
)

func TestParseCPUinfoKStats(t *testing.T) {
	t.Parallel()

	testData := sth.FromFixture("cpu_info:0:cpu_info0.kstat")
	fields, tags := parseCPUinfoKStats(testData)

	require.Equal(
		t,
		map[string]interface{}{
			"speed": float64(2701000000),
		},
		fields,
	)

	require.Equal(
		t,
		map[string]string{
			"coreID":   "0",
			"chipID":   "0",
			"state":    "on-line",
			"clockMHz": "2712",
		},
		tags,
	)
}

func TestParseZoneCPUKStats(t *testing.T) {
	t.Parallel()

	testData := sth.FromFixture("zones:5:cube-ws.kstat")
	fields, tags := parseZoneCPUKStats(testData)

	require.Equal(
		t,
		map[string]interface{}{
			"sys":  float64(1971772437360),
			"user": float64(227321585182107),
		},
		fields,
	)

	require.Equal(
		t,
		map[string]string{
			"name": "cube-ws",
		},
		tags,
	)
}

func TestParseSysCPUKStats(t *testing.T) {
	t.Parallel()

	s := &IllumosCPU{
		SysFields: []string{"cpu_nsec_dtrace", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"},
	}

	testData := sth.FromFixture("cpu:3:sys.kstat")
	fields := parseSysCPUKStats(s, testData)

	require.Equal(
		t,
		map[string]interface{}{
			"nsec.dtrace": float64(4116862855),
			"nsec.intr":   float64(3203901292813),
			"nsec.kernel": float64(31203774025847),
			"nsec.user":   float64(105926508484786),
		},
		fields,
	)
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosCPU{
		CPUInfoStats: true,
		ZoneCPUStats: true,
		SysFields:    []string{"cpu_nsec_kernel", "cpu_nsec_user"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	for _, metric := range acc.GetTelegrafMetrics() {
		switch metric.Name() {
		case "cpu.info":
			testCPUinfoMetric(t, metric)
		case "cpu.zone":
			testZoneCPUMetric(t, metric)
		case "cpu":
			testSysMetric(t, metric)
		}
	}
}

func testCPUinfoMetric(t *testing.T, metric telegraf.Metric) {
	t.Helper()
	require.True(t, metric.HasField("speed"))
	require.True(t, metric.HasTag("coreID"))
	require.True(t, metric.HasTag("chipID"))
	require.True(t, metric.HasTag("state"))
	require.True(t, metric.HasTag("clockMHz"))
}

func testZoneCPUMetric(t *testing.T, metric telegraf.Metric) {
	t.Helper()
	require.True(t, metric.HasField("sys"))
	require.True(t, metric.HasField("user"))
	require.True(t, metric.HasTag("name"))
}

func testSysMetric(t *testing.T, metric telegraf.Metric) {
	t.Helper()
	require.True(t, metric.HasField("nsec.kernel"))
	require.True(t, metric.HasField("nsec.user"))
	require.True(t, metric.HasTag("coreID"))
}
