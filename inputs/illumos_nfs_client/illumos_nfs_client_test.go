package illumos_nfs_client

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// This test is sketchy. It needs to run on a system with kstats, and worse than that, it needs to
// run on a system which hasn't served and NFS content. I imagine it's possible to mock the kstat
// calls, but it's something I don't have the energy for at the moment.
func TestPlugin(t *testing.T) {
	s := &IllumosNfsClient{
		Fields:      []string{"read", "write", "remove", "create"},
		NfsVersions: []string{"v3", "v4"},
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
		"nfs.client",
		map[string]string{
			"nfsVersion": "v3",
		},
		map[string]interface{}{
			"create": uint64(0),
			"write":  uint64(0),
			"remove": uint64(0),
			"read":   uint64(0),
		},
		time.Now(),
	),
	testutil.MustMetric(
		"nfs.client",
		map[string]string{
			"nfsVersion": "v4",
		},
		map[string]interface{}{
			"create": uint64(0),
			"write":  uint64(0),
			"remove": uint64(0),
			"read":   uint64(0),
		},
		time.Now(),
	),
}
