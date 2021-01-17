package rqlite

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/rqlite/rqlite/auth"
	"github.com/rqlite/rqlite/cluster"
	"github.com/rqlite/rqlite/disco"
	httpd "github.com/rqlite/rqlite/http"
	"github.com/rqlite/rqlite/store"
	"github.com/rqlite/rqlite/tcp"
)

const logo = `
            _ _ _
           | (_) |
  _ __ __ _| |_| |_ ___
 | '__/ _  | | | __/ _ \   The lightweight, distributed
 | | | (_| | | | ||  __/   relational database.
 |_|  \__, |_|_|\__\___|
         | |               www.rqlite.com
         |_|
`

// These variables are populated via the Go linker.
var (
	version   = "5"
	commit    = "unknown"
	branch    = "unknown"
	buildtime = "unknown"
	features  = []string{}
)

const name = `rqlited`
const desc = `rqlite is a lightweight, distributed relational database, which uses SQLite as its
storage engine. It provides an easy-to-use, fault-tolerant store for relational data.`

func init() {

}

func Start() {
	if config.RqliteParams.ShowVersion {
		fmt.Printf("%s %s %s %s %s (commit %s, branch %s)\n",
			name, version, runtime.GOOS, runtime.GOARCH, runtime.Version(), commit, branch)
		os.Exit(0)
	}

	// Ensure the data path is set.
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "fatal: no data directory set\n")
		os.Exit(1)
	}

	// Ensure no args come after the data directory.
	if flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "fatal: arguments after data directory are not accepted\n")
		os.Exit(1)
	}

	dataPath := flag.Arg(0)

	// Display logo.
	fmt.Println(logo)

	// Configure logging and pump out initial message.
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
	log.SetPrefix(fmt.Sprintf("[%s] ", name))
	log.Printf("%s starting, version %s, commit %s, branch %s", name, version, commit, branch)
	log.Printf("%s, target architecture is %s, operating system target is %s", runtime.Version(), runtime.GOARCH, runtime.GOOS)

	// Start requested profiling.
	startProfile(config.RqliteParams.CpuProfile, config.RqliteParams.MemProfile)

	// Create internode network layer.
	var tn *tcp.Transport
	if config.RqliteParams.NodeEncrypt {
		log.Printf("enabling node-to-node encryption with cert: %s, key: %s", config.RqliteParams.NodeX509Cert, config.RqliteParams.NodeX509Key)
		tn = tcp.NewTLSTransport(config.RqliteParams.NodeX509Cert, config.RqliteParams.NodeX509Key, config.RqliteParams.NoVerify)
	} else {
		tn = tcp.NewTransport()
	}
	if err := tn.Open(config.RqliteParams.RaftAddr); err != nil {
		log.Fatalf("failed to open internode network layer: %s", err.Error())
	}

	// Create and open the store.
	dataPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Fatalf("failed to determine absolute data path: %s", err.Error())
	}
	dbConf := store.NewDBConfig(config.RqliteParams.Dsn, !config.RqliteParams.OnDisk)

	str := store.New(tn, &store.StoreConfig{
		DBConf: dbConf,
		Dir:    dataPath,
		ID:     idOrRaftAddr(),
	})

	// Set optional parameters on store.
	str.SetRequestCompression(config.RqliteParams.CompressionBatch, config.RqliteParams.CompressionSize)
	str.RaftLogLevel = config.RqliteParams.RaftLogLevel
	str.ShutdownOnRemove = config.RqliteParams.RaftShutdownOnRemove
	str.SnapshotThreshold = config.RqliteParams.RaftSnapThreshold
	str.SnapshotInterval, err = time.ParseDuration(config.RqliteParams.RaftSnapInterval)
	if err != nil {
		log.Fatalf("failed to parse Raft Snapsnot interval %s: %s", config.RqliteParams.RaftSnapInterval, err.Error())
	}
	str.LeaderLeaseTimeout, err = time.ParseDuration(config.RqliteParams.RaftLeaderLeaseTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft Leader lease timeout %s: %s", config.RqliteParams.RaftLeaderLeaseTimeout, err.Error())
	}
	str.HeartbeatTimeout, err = time.ParseDuration(config.RqliteParams.RaftHeartbeatTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft heartbeat timeout %s: %s", config.RqliteParams.RaftHeartbeatTimeout, err.Error())
	}
	str.ElectionTimeout, err = time.ParseDuration(config.RqliteParams.RaftElectionTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft election timeout %s: %s", config.RqliteParams.RaftElectionTimeout, err.Error())
	}
	str.ApplyTimeout, err = time.ParseDuration(config.RqliteParams.RaftApplyTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft apply timeout %s: %s", config.RqliteParams.RaftApplyTimeout, err.Error())
	}

	// Any prexisting node state?
	var enableBootstrap bool
	isNew := store.IsNewNode(dataPath)
	if isNew {
		log.Printf("no preexisting node state detected in %s, node may be bootstrapping", dataPath)
		enableBootstrap = true // New node, so we may be bootstrapping
	} else {
		log.Printf("preexisting node state detected in %s", dataPath)
	}

	// Determine join addresses
	var joins []string
	joins, err = determineJoinAddresses()
	if err != nil {
		log.Fatalf("unable to determine join addresses: %s", err.Error())
	}

	// Supplying join addresses means bootstrapping a new cluster won't
	// be required.
	if len(joins) > 0 {
		enableBootstrap = false
		log.Println("join addresses specified, node is not bootstrapping")
	} else {
		log.Println("no join addresses set")
	}

	// Join address supplied, but we don't need them!
	if !isNew && len(joins) > 0 {
		log.Println("node is already member of cluster, ignoring join addresses")
	}

	// Now, open store.
	if err := str.Open(enableBootstrap); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	// Prepare metadata for join command.
	apiAdv := config.RqliteParams.HttpAddr
	if config.RqliteParams.HttpAdv != "" {
		apiAdv = config.RqliteParams.HttpAdv
	}
	apiProto := "http"
	if config.RqliteParams.X509Cert != "" {
		apiProto = "https"
	}
	meta := map[string]string{
		"api_addr":  apiAdv,
		"api_proto": apiProto,
	}

	// Execute any requested join operation.
	if len(joins) > 0 && isNew {
		log.Println("join addresses are:", joins)
		advAddr := config.RqliteParams.RaftAddr
		if config.RqliteParams.RaftAdv != "" {
			advAddr = config.RqliteParams.RaftAdv
		}

		joinDur, err := time.ParseDuration(config.RqliteParams.JoinInterval)
		if err != nil {
			log.Fatalf("failed to parse Join interval %s: %s", config.RqliteParams.JoinInterval, err.Error())
		}

		tlsConfig := tls.Config{InsecureSkipVerify: config.RqliteParams.NoVerify}
		if config.RqliteParams.X509CACert != "" {
			asn1Data, err := ioutil.ReadFile(config.RqliteParams.X509CACert)
			if err != nil {
				log.Fatalf("ioutil.ReadFile failed: %s", err.Error())
			}
			tlsConfig.RootCAs = x509.NewCertPool()
			ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(asn1Data))
			if !ok {
				log.Fatalf("failed to parse root CA certificate(s) in %q", config.RqliteParams.X509CACert)
			}
		}

		if j, err := cluster.Join(joins, str.ID(), advAddr, !config.RqliteParams.RaftNonVoter, meta,
			config.RqliteParams.JoinAttempts, joinDur, &tlsConfig); err != nil {
			log.Fatalf("failed to join cluster at %s: %s", joins, err.Error())
		} else {
			log.Println("successfully joined cluster at", j)
		}

	}

	// Wait until the store is in full consensus.
	openTimeout, err := time.ParseDuration(config.RqliteParams.RaftOpenTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft open timeout %s: %s", config.RqliteParams.RaftOpenTimeout, err.Error())
	}
	str.WaitForLeader(openTimeout)
	str.WaitForApplied(openTimeout)

	// This may be a standalone server. In that case set its own metadata.
	if err := str.SetMetadata(meta); err != nil && err != store.ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		log.Fatalf("failed to set store metadata: %s", err.Error())
	}

	// Start the HTTP API server.
	if err := startHTTPService(str); err != nil {
		log.Fatalf("failed to start HTTP server: %s", err.Error())
	}
	log.Println("node is ready")

	// Block until signalled.
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	if err := str.Close(true); err != nil {
		log.Printf("failed to close store: %s", err.Error())
	}
	stopProfile()
	log.Println("rqlite server stopped")
}

func determineJoinAddresses() ([]string, error) {
	apiAdv := config.RqliteParams.HttpAddr
	if config.RqliteParams.HttpAdv != "" {
		apiAdv = config.RqliteParams.HttpAdv
	}

	var addrs []string
	if config.RqliteParams.JoinAddr != "" {
		// Explicit join addresses are first priority.
		addrs = strings.Split(config.RqliteParams.JoinAddr, ",")
	}

	if config.RqliteParams.DiscoID != "" {
		log.Printf("registering with Discovery Service at %s with ID %s", config.RqliteParams.DiscoURL, config.RqliteParams.DiscoID)
		c := disco.New(config.RqliteParams.DiscoURL)
		r, err := c.Register(config.RqliteParams.DiscoID, apiAdv)
		if err != nil {
			return nil, err
		}
		log.Println("Discovery Service responded with nodes:", r.Nodes)
		for _, a := range r.Nodes {
			if a != apiAdv {
				// Only other nodes can be joined.
				addrs = append(addrs, a)
			}
		}
	}

	return addrs, nil
}

func startHTTPService(str *store.Store) error {
	// Get the credential store.
	credStr, err := credentialStore()
	if err != nil {
		return err
	}

	// Create HTTP server and load authentication information if required.
	var s *httpd.Service
	if credStr != nil {
		s = httpd.New(config.RqliteParams.HttpAddr, str, credStr)
	} else {
		s = httpd.New(config.RqliteParams.HttpAddr, str, nil)
	}

	s.CertFile = config.RqliteParams.X509Cert
	s.KeyFile = config.RqliteParams.X509Key
	s.Expvar = config.RqliteParams.Expvar
	s.Pprof = config.RqliteParams.PprofEnabled
	s.BuildInfo = map[string]interface{}{
		"commit":     commit,
		"branch":     branch,
		"version":    version,
		"build_time": buildtime,
	}
	return s.Start()
}

func credentialStore() (*auth.CredentialsStore, error) {
	if config.RqliteParams.AuthFile == "" {
		return nil, nil
	}

	f, err := os.Open(config.RqliteParams.AuthFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open authentication file %s: %s", config.RqliteParams.AuthFile, err.Error())
	}

	cs := auth.NewCredentialsStore()
	if cs.Load(f); err != nil {
		return nil, err
	}
	return cs, nil
}

func idOrRaftAddr() string {
	if config.RqliteParams.NodeID != "" {
		return config.RqliteParams.NodeID
	}
	if config.RqliteParams.RaftAdv == "" {
		return config.RqliteParams.RaftAddr
	}
	return config.RqliteParams.RaftAdv
}

// prof stores the file locations of active profiles.
var prof struct {
	cpu *os.File
	mem *os.File
}

// startProfile initializes the CPU and memory profile, if specified.
func startProfile(cpuprofile, memprofile string) {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("failed to create CPU profile file at %s: %s", cpuprofile, err.Error())
		}
		log.Printf("writing CPU profile to: %s\n", cpuprofile)
		prof.cpu = f
		pprof.StartCPUProfile(prof.cpu)
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("failed to create memory profile file at %s: %s", cpuprofile, err.Error())
		}
		log.Printf("writing memory profile to: %s\n", memprofile)
		prof.mem = f
		runtime.MemProfileRate = 4096
	}
}

// stopProfile closes the CPU and memory profiles if they are running.
func stopProfile() {
	if prof.cpu != nil {
		pprof.StopCPUProfile()
		prof.cpu.Close()
		log.Println("CPU profiling stopped")
	}
	if prof.mem != nil {
		pprof.Lookup("heap").WriteTo(prof.mem, 0)
		prof.mem.Close()
		log.Println("memory profiling stopped")
	}
}
