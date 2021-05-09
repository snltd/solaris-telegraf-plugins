package illumos_proc

import (
	//"encoding/binary"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	//sh "github.com/snltd/solaris-telegraf-helpers"
	//"io/ioutil"
	//"github.com/fatih/structs"
	//"log"
	//"os"
	//"sort"
	//"strconv"
	//"strings"
)

var sampleConfig = `
	## Everything in 'Fields' will create a new metric path, like "proc.<execname>.<field>".
	# fields = ["size", "rssize"]
	## How many processes to send metrics for.
	# top_n = 10
	## Which tags to apply. Some, like the SMF service, are a little expensive.
	# tags = ["name", "pid", "zone", "svc"]
`

func (s *IllumosProc) Description() string {
	return "Reports on Illumos processes, like prstat(1)"
}

func (s *IllumosProc) SampleConfig() string {
	return sampleConfig
}

type IllumosProc struct {
	Fields []string
	Tags   []string
	TopN   int
}

func (s *IllumosProc) Gather(acc telegraf.Accumulator) error {
	fmt.Printf("merp\n")
	return nil
}

/*
type procDigest struct {
	pid   int    // process ID
	name  string // exec name
	value int64  // value of counter
	zone  string // zone name
	ctid  id_t   // contract ID (for SMF service name lookup)
	ts    int64  // nanosecond time stamp, relative to global boot
}

var procDir = "/proc"
var zoneMap sh.ZoneMap

// leaderboard() needs these procDigest structs and funcs to sort
type procDigests []procDigest
type procItems map[string]interface{}

func (d procDigests) Len() int           { return len(d) }
func (d procDigests) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d procDigests) Less(i, j int) bool { return d[i].value < d[j].value }

func procPidList() []int {
	procs, err := os.ReadDir(procDir)

	if err != nil {
		log.Fatal(fmt.Sprintf("cannot read %v", procDir))
	}

	var ret []int

	for _, proc := range procs {
		pid, _ := strconv.Atoi(proc.Name())
		ret = append(ret, pid)
	}

	return ret
}

/////////////////////////////////

func allProcs() map[int]procItems {
	procs := procPidList()
	ret := make(map[int]procItems)

	for _, pid := range procs {
		psinfo, info_err := procPsinfo(pid)

		if info_err == nil {
			m := structs.Map(psinfo)

			usage, usage_err := procUsage(pid)

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

var runSvcsCtidCmd = func() string {
	return sh.RunCmd("/bin/svcs -vHo ctid,fmri")
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

func zoneLookup(zid id_t) string {
	zone, _ := zoneMap.ZoneByID(int(zid))
	return zone.Name
}

// Return the service name associated with a contract Id. If there
// isn't one, it'll return the empty string, which is fine.
//
// func ctidToSvc(ctmap map[id_t]string, ctid id_t) string {
	// svc := ctmap[ctid]
	// return svc
// }

func (s *IllumosProc) Gather(acc telegraf.Accumulator) error {
	all_procs := allProcs()
	var contract_map map[int]string

	if sh.WeWant("svc", s.Tags) {
		contract_map = contractMap(runSvcsCtidCmd())
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
				tags["svc"] = contract_map[int(proc.ctid)]
			}

			metrics[field] = proc.value
			acc.AddFields("solaris_proc", metrics, tags)
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////

func contractMap(svcsOutput string) map[int]string {
	ret := make(map[int]string)

	for _, svcLine := range strings.Split(svcsOutput, "\n") {
		fields := strings.Fields(svcLine)
		ctidStr := fields[0]
		svc := fields[1]

		if ctidStr == "-" {
			continue
		}

		ctid, err := strconv.Atoi(fields[0])

		if err == nil {
			ret[ctid] = svc
		}
	}

	return ret
}

func procPsinfo(pid int) (psinfo_t, error) {
	var psinfo psinfo_t

	file := fmt.Sprintf("%s/%d/psinfo", procDir, pid)
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

func procUsage(pid int) (prusage_t, error) {
	file := fmt.Sprintf("%s/%d/usage", procDir, pid)
	var prusage prusage_t

	fh, err := os.Open(file)

	if err != nil {
		log.Printf("cannot open %s\n", file)
		return prusage, err
	}

	err = binary.Read(fh, binary.LittleEndian, &prusage)

	if err != nil {
		log.Printf("cannot read %s\n", file)
		return prusage, err
	}

	return prusage, err
}

*/
func init() {
	//zoneMap = sh.NewZoneMap()
	inputs.Add("illumos_proc", func() telegraf.Input { return &IllumosProc{} })
}
