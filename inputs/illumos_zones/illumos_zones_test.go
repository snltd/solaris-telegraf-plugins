package illumos_zones

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPlugin(t *testing.T) {
	s := &IllumosZones{}

	makeZoneMap = func() sth.ZoneMap {
		return sth.ParseZones(zoneadmOutput)
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	testutil.RequireMetricsEqual(
		t,
		testMetrics,
		acc.GetTelegrafMetrics(),
		testutil.SortMetrics(),
		testutil.IgnoreTime())
}

var zoneadmOutput = `0:global:running:/::ipkg:shared:0
42:cube-media:running:/zones/cube-media:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128
44:cube-ws:installed:/zones/cube-ws:0f9c56f4-9810-6d45-f801-d34bf27cc13f:pkgsrc:excl:179`

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "running",
			"ipType": "shared",
			"brand":  "ipkg",
			"name":   "global",
		},
		map[string]interface{}{
			"status": 1,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "installed",
			"ipType": "excl",
			"brand":  "pkgsrc",
			"name":   "cube-ws",
		},
		map[string]interface{}{
			"status": 0,
		},
		time.Now(),
	),
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "running",
			"ipType": "excl",
			"brand":  "lipkg",
			"name":   "cube-media",
		},
		map[string]interface{}{
			"status": 1,
		},
		time.Now(),
	),
}
