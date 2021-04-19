package illumos_zones

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPlugin(t *testing.T) {
	s := &IllumosZones{
		ZoneStates: []string{"running", "installed", "incomplete"},
		ZoneBrands: []string{"ipkg", "lipkg", "sparse"},
		ZoneID:     true,
		ZoneCount:  true,
	}

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

func TestZeroedMap(t *testing.T) {
	assert.Equal(t,
		map[string]int{
			"configured": 0,
			"running":    0,
		},
		zeroedMap([]string{"running", "configured"}),
	)
}

func TestGatherData(t *testing.T) {
	s := &IllumosZones{
		ZoneStates: []string{"running", "installed"},
		ZoneBrands: []string{"ipkg", "lipkg"},
	}

	assert.Equal(
		t,
		map[string]interface{}{
			"brand.ipkg":      1,
			"brand.lipkg":     1,
			"state.installed": 1,
			"state.running":   2,
		},
		gatherCounts(s, sth.ParseZones(zoneadmOutput)))
}

var zoneadmOutput = `0:global:running:/::ipkg:shared:0
42:cube-media:running:/zones/cube-media:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128
44:cube-ws:installed:/zones/cube-ws:0f9c56f4-9810-6d45-f801-d34bf27cc13f:pkgsrc:excl:179`

var testMetrics = []telegraf.Metric{
	testutil.MustMetric(
		"zones",             // name
		map[string]string{}, //tags
		map[string]interface{}{ // fields
			"state.running":    2,
			"state.installed":  1,
			"state.incomplete": 0,
			"brand.ipkg":       1,
			"brand.lipkg":      1,
			"brand.sparse":     0,
			"count":            3,
		},
		time.Now(), //time
	),
	testutil.MustMetric(
		"zones",
		map[string]string{
			"status": "running",
			"ipType": "shared",
			"brand":  "ipkg",
			"name":   "global",
		},
		map[string]interface{}{
			"properties": 1,
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
			"properties": 0,
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
			"properties": 1,
		},
		time.Now(),
	),
}
