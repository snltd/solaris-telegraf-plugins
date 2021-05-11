package illumos_smf

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosSmf{
		SvcStates:       []string{"online", "maintenance"},
		Zones:           []string{"global", "cube-pkgsrc"},
		GenerateDetails: true,
	}

	rawOutput = func() string {
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
		"smf",
		map[string]string{
			"zone":  "cube-pkgsrc",
			"state": "online",
		},
		map[string]interface{}{
			"states": 4,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"smf",
		map[string]string{
			"zone":  "cube-pkgsrc",
			"state": "maintenance",
		},
		map[string]interface{}{
			"states": 1,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"smf",
		map[string]string{
			"zone":  "global",
			"state": "online",
		},
		map[string]interface{}{
			"states": 2,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"smf",
		map[string]string{
			"zone":  "cube-pkgsrc",
			"state": "maintenance",
			"fmri":  "svc:/system/filesystem/local:default",
		},
		map[string]interface{}{
			"errors": 1,
		},
		time.Now(),
	),
}

func TestParseSvcsNoFilters(t *testing.T) {
	t.Parallel()

	testConfig := IllumosSmf{}

	assert.Equal(
		t,
		svcSummary{
			counts: svcCounts{
				"cube-pkgsrc": zoneSvcSummary{
					"online":      4,
					"maintenance": 1,
					"disabled":    1,
				},
				"cube-cron": zoneSvcSummary{
					"legacy_run": 2,
					"online":     3,
					"disabled":   2,
				},
				"global": zoneSvcSummary{
					"legacy_run": 3,
					"online":     2,
					"disabled":   1,
				},
			},
			svcErrs: svcErrs{},
		},
		parseSvcs(testConfig, sampleOutput))
}

func TestParseSvcsFilters(t *testing.T) {
	t.Parallel()

	testConfig := IllumosSmf{
		SvcStates:       []string{"online", "maintenance"},
		Zones:           []string{"global", "cube-pkgsrc"},
		GenerateDetails: true,
	}

	assert.Equal(
		t,
		svcSummary{
			counts: svcCounts{
				"cube-pkgsrc": zoneSvcSummary{
					"online":      4,
					"maintenance": 1,
				},
				"global": zoneSvcSummary{
					"online": 2,
				},
			},

			svcErrs: svcErrs{
				svcErr{
					zone:  "cube-pkgsrc",
					state: "maintenance",
					fmri:  "svc:/system/filesystem/local:default",
				},
			},
		},
		parseSvcs(testConfig, sampleOutput))
}

var sampleOutput = `cube-pkgsrc      maintenance    svc:/system/filesystem/local:default
cube-pkgsrc      online         svc:/system/filesystem/minimal:default
cube-pkgsrc      online         svc:/system/manifest-import:default
cube-pkgsrc      online         svc:/system/identity:node
cube-pkgsrc      online         svc:/system/boot-archive:default
cube-pkgsrc      disabled       svc:/system/svc/global:default
cube-cron        legacy_run     lrc:/etc/rc2_d/S89PRESERVE
cube-cron        legacy_run     lrc:/etc/rc2_d/S20sysetup
cube-cron        online         svc:/sysdef/puppet:default
cube-cron        online         svc:/system/boot-config:default
cube-cron        online         svc:/system/device/audio:default
cube-cron        disabled       svc:/system/device/mpxio-upgrade:default
cube-cron        disabled       svc:/system/device/allocate:default
global           legacy_run     lrc:/etc/rc2_d/S89PRESERVE
global           legacy_run     lrc:/etc/rc2_d/S81dodatadm_udaplt
global           legacy_run     lrc:/etc/rc2_d/S20sysetup
global           online         svc:/system/config-assemble:services
global           online         svc:/sdef/diamond:default
global           disabled       svc:/network/varpd:default`
