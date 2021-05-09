package illumos_nfs_client

import (
	"github.com/influxdata/telegraf/testutil"
	sth "github.com/snltd/solaris-telegraf-helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// The meat of the plugin is tested by TestParseNamedStats. This exercises the full code path,
// hittng real kstats
func TestPlugin(t *testing.T) {
	s := &IllumosNfsClient{
		Fields:      []string{"read", "write", "remove", "create"},
		NfsVersions: []string{"v4"},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, s.Gather(&acc))
	metric := acc.GetTelegrafMetrics()[0]

	assert.Equal(t, "nfs.client", metric.Name())
	assert.True(t, metric.HasTag("nfsVersion"))

	for _, field := range s.Fields {
		_, present := metric.GetField(field)
		assert.True(t, present)
	}
}

func TestParseNamedStats(t *testing.T) {
	s := &IllumosNfsClient{
		Fields:      []string{"read", "write", "remove", "create"},
		NfsVersions: []string{"v4"},
	}

	testData := sth.FromFixture("nfs:0:rfsreqcnt_v4.kstat")
	fields := parseNamedStats(s, testData)

	assert.Equal(
		t,
		fields,
		map[string]interface{}{
			"read":   float64(23010),
			"write":  float64(750),
			"remove": float64(317),
			"create": float64(3),
		},
	)
}
