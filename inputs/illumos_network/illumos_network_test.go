package illumos_network

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNamedStats(t *testing.T) {
	t.Parallel()

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64", "ipackets64"},
		Zones:  []string{"cube-dns"},
	}

	testData := sth.FromFixture("link:0:dns_net0.kstat")
	fields := parseNamedStats(s, testData)

	assert.Equal(
		t,
		map[string]interface{}{
			"obytes64":   float64(69053870),
			"rbytes64":   float64(1518773044),
			"ipackets64": float64(1637072),
		},
		fields,
	)
}

func TestParseNamedStatsNoSelectedNics(t *testing.T) {
	t.Parallel()

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64", "ipackets64"},
		Zones:  []string{"cube-dns"},
		Vnics:  []string{"net0"},
	}

	testData := sth.FromFixture("link:0:dns_net0.kstat")
	fields := parseNamedStats(s, testData)
	assert.Equal(t, map[string]interface{}{}, fields)
}

func TestZoneTags(t *testing.T) {
	t.Parallel()

	zoneName = "global"

	assert.Equal(
		t,
		map[string]string{
			"zone":  "cube-dns",
			"link":  "rge0",
			"speed": "1000mbit",
			"name":  "dns_net0",
		},
		zoneTags("cube-dns", "dns_net0", sth.ParseZoneVnics(sampleDladmOutput)["dns_net0"]),
	)
}

func TestZoneTagsGlobal(t *testing.T) {
	t.Parallel()

	zoneName = "global"

	assert.Equal(
		t,
		map[string]string{
			"zone":  "global",
			"link":  "none",
			"speed": "unknown",
			"name":  "rge0",
		},
		zoneTags("global", "rge0", sth.ParseZoneVnics(sampleDladmOutput)["rge0"]),
	)
}

func TestPlugin(t *testing.T) {
	t.Parallel()

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64", "collisions", "ierrors"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))

	metric := acc.GetTelegrafMetrics()[0]
	assert.Equal(t, "net", metric.Name())
	assert.True(t, metric.HasTag("zone"))
	assert.True(t, metric.HasTag("link"))
	assert.True(t, metric.HasTag("speed"))
	assert.True(t, metric.HasTag("name"))

	for _, field := range s.Fields {
		_, present := metric.GetField(field)
		assert.True(t, present)
	}
}

var sampleDladmOutput = `media_net0:cube-media:rge0:1000
dns_net0:cube-dns:rge0:1000
pkgsrc_net0:cube-pkgsrc:rge0:1000
backup_net0:cube-backup:rge0:1000`
