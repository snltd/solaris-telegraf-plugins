/*
At the moment this collector sends everything as a gauge, so it's up
to you and your graphing software to convert them into meaningful rates.
It's work-in-progress, rough-and-ready, there to do a quick job.

If you want to add more tags, like project ID or something, they
need to go in the procDigest struct.

*/

package solaris_proc

import (
	"encoding/binary"
	"fmt"
	"github.com/fatih/structs"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	sh "github.com/snltd/sunos_telegraf_helpers"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var sampleConfig = `
	## Everything in 'Fields' will create a new metric path, as in
	## "proc.<execname>.<field>".
	# Fields = ["size", "rssize"]
	## How many processes to send metrics for.
	# TopN = 10
	## Which tags to apply. Some, like the SMF service, are a little
	## expensive
	# Tags = ["name", "pid", "zone", "svc"]
`

func (s *SolarisProc) Description() string {
	return "Reports on Solaris processes, like prstat(1)"
}

func (s *SolarisProc) SampleConfig() string {
	return sampleConfig
}

type procDigest struct {
	pid   int    // process ID
	name  string // exec name
	value int64  // value of counter
	zone  string // zone name
	ctid  id_t   // contract ID (for SMF service name lookup)
	ts    int64  // nanosecond time stamp, relative to global boot
}

// leaderboard() needs to sort
//
type procDigests []procDigest
type procItems map[string]interface{}

func (d procDigests) Len() int           { return len(d) }
func (d procDigests) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d procDigests) Less(i, j int) bool { return d[i].value < d[j].value }

type SolarisProc struct {
	Fields []string
	Tags   []string
	TopN   int
}

// The following types come from /usr/include/sys/procfs.h, with thanks
// to https://github.com/mitchellh/go-ps/blob/master/process_solaris.go
// for getting me started

type ushort_t uint16
type id_t int32
type pid_t int32
type uid_t int32
type gid_t int32
type dev_t uint64
type size_t uint64
type uintptr_t uint64
type ulong_t uint64
type timestruc_t [2]int64

type prusage_t struct {
	Pr_lwpid    id_t           /* lwp id.  0: process or defunct */
	Pr_count    int32          /* number of contributing lwps */
	Pr_tstamp   timestruc_t    /* current time stamp */
	Pr_create   timestruc_t    /* process/lwp creation time stamp */
	Pr_term     timestruc_t    /* process/lwp termination time stamp */
	Pr_rtime    timestruc_t    /* total lwp real (elapsed) time */
	Pr_utime    timestruc_t    /* user level cpu time */
	Pr_stime    timestruc_t    /* system call cpu time */
	Pr_ttime    timestruc_t    /* other system trap cpu time */
	Pr_tftime   timestruc_t    /* text page fault sleep time */
	Pr_dftime   timestruc_t    /* data page fault sleep time */
	Pr_kftime   timestruc_t    /* kernel page fault sleep time */
	Pr_ltime    timestruc_t    /* user lock wait sleep time */
	Pr_slptime  timestruc_t    /* all other sleep time */
	Pr_wtime    timestruc_t    /* wait-cpu (latency) time */
	Pr_stoptime timestruc_t    /* stopped time */
	Filltime    [6]timestruc_t /* filler for future expansion */
	Pr_minf     ulong_t        /* minor page faults */
	Pr_majf     ulong_t        /* major page faults */
	Pr_nswap    ulong_t        /* swaps */
	Pr_inblk    ulong_t        /* input blocks */
	Pr_oublk    ulong_t        /* output blocks */
	Pr_msnd     ulong_t        /* messages sent */
	Pr_mrcv     ulong_t        /* messages received */
	Pr_sigs     ulong_t        /* signals received */
	Pr_vctx     ulong_t        /* voluntary context switches */
	Pr_ictx     ulong_t        /* involuntary context switches */
	Pr_sysc     ulong_t        /* system calls */
	Pr_ioch     ulong_t        /* chars read and written */
	Filler      [10]ulong_t    /* filler for future expansion */
}

type psinfo_t struct {
	Pr_flag   int32     /* process flags (DEPRECATED; do not use) */
	Pr_nlwp   int32     /* number of active lwps in the process */
	Pr_pid    pid_t     /* unique process id */
	Pr_ppid   pid_t     /* process id of parent */
	Pr_pgid   pid_t     /* pid of process group leader */
	Pr_sid    pid_t     /* session id */
	Pr_uid    uid_t     /* real user id */
	Pr_euid   uid_t     /* effective user id */
	Pr_gid    gid_t     /* real group id */
	Pr_egid   gid_t     /* effective group id */
	Pr_addr   uintptr_t /* address of process */
	Pr_size   size_t    /* size of process image in Kbytes */
	Pr_rssize size_t    /* resident set size in Kbytes */
	Pr_pad1   size_t
	Pr_ttydev dev_t /* controlling tty device (or PRNODEV) */

	// Guess this following 2 ushort_t values require a padding to properly
	// align to the 64bit mark.
	Pr_pctcpu   ushort_t /* % of recent cpu time used by all lwps */
	Pr_pctmem   ushort_t /* % of system memory used by process */
	Pr_pad64bit [4]byte

	Pr_start    timestruc_t /* process start time, from the epoch */
	Pr_time     timestruc_t /* usr+sys cpu time for this process */
	Pr_ctime    timestruc_t /* usr+sys cpu time for reaped children */
	Pr_fname    [16]byte    /* name of execed file */
	Pr_psargs   [80]byte    /* initial characters of arg list */
	Pr_wstat    int32       /* if zombie, the wait() status */
	Pr_argc     int32       /* initial argument count */
	Pr_argv     uintptr_t   /* address of initial argument vector */
	Pr_envp     uintptr_t   /* address of initial environment vector */
	Pr_dmodel   [1]byte     /* data model of the process */
	Pr_pad2     [3]byte
	Pr_taskid   id_t      /* task id */
	Pr_projid   id_t      /* project id */
	Pr_nzomb    int32     /* number of zombie lwps in the process */
	Pr_poolid   id_t      /* pool id */
	Pr_zoneid   id_t      /* zone id */
	Pr_contract id_t      /* process contract */
	Pr_filler   int32     /* reserved for future use */
	Pr_lwp      [128]byte /* information for representative lwp */
}

func all_procs() map[int]procItems {
	procs, err := ioutil.ReadDir("/proc")

	if err != nil {
		log.Fatal("cannot read /proc")
	}

	ret := make(map[int]procItems)

	for _, proc := range procs {
		pid, _ := strconv.Atoi(proc.Name())
		psinfo, info_err := proc_psinfo(pid)

		if info_err == nil {
			m := structs.Map(psinfo)

			usage, usage_err := proc_usage(pid)

			if usage_err == nil {
				n := structs.Map(usage)

				for k, v := range n {
					m[k] = v
				}

				ret[pid] = m
			}
		}

	}

	return ret
}

func proc_usage(pid int) (prusage_t, error) {
	file := fmt.Sprintf("/proc/%d/usage", pid)
	var prusage prusage_t

	fh, err := os.Open(file)

	if err != nil {
		log.Printf("cannot open %s\n", file)
		return prusage, err
	}

	err = binary.Read(fh, binary.LittleEndian, &prusage)

	if err != nil {
		//fmt.Printf("%+v\n", &prusage)
		log.Printf("cannot read %s\n", file)
		log.Printf("%s\n", err)
		log.Printf("%T\n", err)

		return prusage, err
	}

	return prusage, err
}

func proc_psinfo(pid int) (psinfo_t, error) {
	file := fmt.Sprintf("/proc/%d/psinfo", pid)
	var psinfo psinfo_t

	fh, err := os.Open(file)

	if err != nil {
		log.Printf("cannot open %s\n", file)
		return psinfo, err
	}

	err = binary.Read(fh, binary.LittleEndian, &psinfo)

	if err != nil {
		log.Printf("cannot read %s\n", file)
		return psinfo, err
	}

	return psinfo, nil
}

// returns a map of contract ID => SMF FMRI
//
func ContractMap() map[id_t]string {
	raw := strings.Split(sh.RunCmd("/bin/svcs -vHo ctid,fmri"), "\n")
	ret := make(map[id_t]string)

	for _, row := range raw {
		fields := strings.Fields(row)
		svc := fields[1]

		if fields[0] != "-" {
			ct, err := strconv.Atoi(fields[0])

			if err == nil {
				ret[id_t(ct)] = svc
			}
		}
	}

	return ret

}

// Returns a list the top 'n' processes, sorted on the field you
// specify
//
func leaderboard(procs map[int]procItems, field string,
	limit int) procDigests {
	var to_sort procDigests

	for pid, vals := range procs {
		// convert the exec name from a byte array
		raw_name := vals["Pr_fname"].([16]byte)
		name := strings.TrimRight(string(raw_name[:]), "\x00")

		// convert timestruc_t into a straight nanosecond value. It'll
		// be easier to work with elsewhere

		raw_ts := vals["Pr_tstamp"].(timestruc_t)
		ts := raw_ts[0]*1e9 + raw_ts[1]

		c := procDigest{
			pid:   pid,
			name:  name,
			value: int64(vals[field].(size_t)),
			zone:  zoneLookup(vals["Pr_zoneid"].(id_t)),
			ctid:  vals["Pr_contract"].(id_t),
			ts:    ts}

		to_sort = append(to_sort, c)
	}

	sort.Sort(procDigests(to_sort))
	sort.Sort(sort.Reverse(to_sort))

	return to_sort[:limit]
}

// Return the zone name with the given ID
//
func zoneLookup(zid id_t) string {
	return sh.ZoneMap()[int(zid)]
}

// Return the service name associated with a contract Id. If there
// isn't one, it'll return the empty string, which is fine.
//
func ctidToSvc(ctmap map[id_t]string, ctid id_t) string {
	svc := ctmap[ctid]
	return svc
}

func (s *SolarisProc) Gather(acc telegraf.Accumulator) error {
	all_procs := all_procs()
	var contract_map map[id_t]string

	if sh.WeWant("svc", s.Tags) {
		contract_map = ContractMap()
	}

	for _, field := range s.Fields {
		raw_field := "Pr_" + field
		procs := leaderboard(all_procs, raw_field, s.TopN)

		for _, proc := range procs {
			metrics := make(map[string]interface{})
			tags := make(map[string]string)

			if sh.WeWant("zone", s.Tags) {
				tags["zone"] = proc.zone
			}

			if sh.WeWant("pid", s.Tags) {
				tags["pid"] = strconv.Itoa(proc.pid)
			}

			if sh.WeWant("name", s.Tags) {
				tags["name"] = proc.name
			}

			if sh.WeWant("svc", s.Tags) {
				tags["svc"] = ctidToSvc(contract_map, proc.ctid)
			}

			metrics[field] = proc.value
			acc.AddFields("solaris_proc", metrics, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("solaris_proc", func() telegraf.Input { return &SolarisProc{} })
}
