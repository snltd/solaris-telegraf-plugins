package illumos_proc

/*
import (
	//"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	//"os"
	"testing"
	"time"
)

func TestPlugin(t *testing.T) {
	s := &IllumosProc{}

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
*/
/*
func TestContractMap(t *testing.T) {
	contractMap := contractMap(sampleSvcsOutput)

	assert.Equal(
		t,
		map[int]string{
			67:  "svc:/network/inetd:default",
			72:  "svc:/network/ssh:default",
			78:  "svc:/network/smb/client:default",
			629: "svc:/sysdef/cron_monitor:default",
			83:  "lrc:/etc/rc2_d/S89PRESERVE",
		},
		contractMap,
	)

	assert.Equal(t, "svc:/network/ssh:default", contractMap[72])
}

// This test uses the current running process to test a few easily accessible members of the
// psinfo_t struct. This means it will only work on Illumos, but so will virtually everything else
// in this repo, so hey-ho.
func TestProcPsinfo(t *testing.T) {
	psinfo, err := procPsinfo(os.Getpid())
	assert.Nil(t, err)
	assert.Equal(t, psinfo.Pr_pid, pid_t(os.Getpid()))
	assert.Equal(t, psinfo.Pr_ppid, pid_t(os.Getppid()))
	assert.Equal(t, psinfo.Pr_uid, uid_t(os.Getuid()))
	assert.Equal(t, psinfo.Pr_gid, gid_t(os.Getgid()))
}

func TestProcPidList(t *testing.T) {
	procDir = "fixtures/proc"
	assert.Equal(t, []int{10887, 11022, 8530}, procPidList())
}

func TestZoneLookup(t *testing.T) {
	assert.Equal(
		t,
		"cube-media",
		zoneLookup(42),
	)

}

var zoneadmOutput = `0:global:running:/::ipkg:shared:0
42:cube-media:running:/zones/cube-media:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128
44:cube-ws:installed:/zones/cube-ws:0f9c56f4-9810-6d45-f801-d34bf27cc13f:pkgsrc:excl:179`

var sampleSvcsOutput = `83 lrc:/etc/rc2_d/S89PRESERVE
   629 svc:/sysdef/cron_monitor:default
     - svc:/sysdef/puppet:default
     - svc:/network/netmask:default
    72 svc:/network/ssh:default
    67 svc:/network/inetd:default
    78 svc:/network/smb/client:default
     - svc:/system/boot-archive:default`
*/
