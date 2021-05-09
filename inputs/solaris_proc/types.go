package illumos_proc

// The following types come from /usr/include/sys/procfs.h, with thanks to
// https://github.com/mitchellh/go-ps/blob/master/process_solaris.go for getting me started

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
	Pr_flag     int32     /* process flags (DEPRECATED; do not use) */
	Pr_nlwp     int32     /* number of active lwps in the process */
	Pr_pid      pid_t     /* unique process id */
	Pr_ppid     pid_t     /* process id of parent */
	Pr_pgid     pid_t     /* pid of process group leader */
	Pr_sid      pid_t     /* session id */
	Pr_uid      uid_t     /* real user id */
	Pr_euid     uid_t     /* effective user id */
	Pr_gid      gid_t     /* real group id */
	Pr_egid     gid_t     /* effective group id */
	Pr_addr     uintptr_t /* address of process */
	Pr_size     size_t    /* size of process image in Kbytes */
	Pr_rssize   size_t    /* resident set size in Kbytes */
	Pr_pad1     size_t
	Pr_ttydev   dev_t    /* controlling tty device (or PRNODEV) */
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
