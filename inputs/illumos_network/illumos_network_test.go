package illumos_network

import (
	"encoding/gob"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/siebenmann/go-kstat"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestPlugin(t *testing.T) {
	kstatData = sampleKstatData

	s := &IllumosNetwork{
		Fields: []string{"obytes64", "rbytes64"},
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
		"net",
		map[string]string{
			"zone": "global",
		},
		map[string]interface{}{
			"rbytes64": uint64(22),
			"obytes64": uint64(10),
		},
		time.Now(),
	),
}

func sampleKstatData() (*kstat.Token, []*kstat.KStat) {
	var kstatData []*kstat.KStat
	var dummyToken *kstat.Token

	raw, err := os.Open("resources/kstat_data")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load serialized data from disk: %v\n", err)
		os.Exit(1)
	}

	dec := gob.NewDecoder(raw)
	err = dec.Decode(&kstatData)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load decode kstat data: %v\n", err)
		os.Exit(1)
	}

	return dummyToken, kstatData
}
