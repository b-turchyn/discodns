package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/jessevdk/go-flags"
	"github.com/miekg/dns"
	"github.com/rcrowley/go-metrics"
)

var (
	logger   = log.New(os.Stderr, "[discodns] ", log.Ldate|log.Ltime)
	logDebug = false

	// Define all of the command line arguments
	options struct {
		ListenAddress    string   `short:"l" long:"listen" description:"Listen IP address" default:"0.0.0.0" env:"DISCODNS_LISTEN_ADDRESS"`
		ListenPort       int      `short:"p" long:"port" description:"Port to listen on" default:"53" env:"DISCODNS_LISTEN_PORT"`
		EtcdHosts        []string `short:"e" long:"etcd" description:"host:port[,host:port] for etcd hosts" default:"127.0.0.1:4001" env:"DISCODNS_ETCD_HOSTS"`
		Debug            bool     `short:"v" long:"debug" description:"Enable debug logging" env:"DISCODNS_DEBUG"`
		MetricsDuration  int      `short:"m" long:"metrics" description:"Dump metrics to stderr every N seconds" default:"30" env:"DISCODNS_METRICS_DURATION"`
		GraphiteServer   string   `long:"graphite" description:"Graphite server to send metrics to" env:"DISCODNS_GRAPHITE_SERVER"`
		GraphiteDuration int      `long:"graphite-duration" description:"Duration to periodically send metrics to the graphite server" default:"10" env:"DISCODNS_GRAPHITE_DURATION"`
		DefaultTTL       uint32   `short:"t" long:"default-ttl" description:"Default TTL to return on records without an explicit TTL" default:"300" env:"DISCODNS_DEFAULT_TTL"`
		Accept           []string `long:"accept" description:"Limit DNS queries to a set of domain:[type,...] pairs" env:"DISCODNS_ACCEPT"`
		Reject           []string `long:"reject" description:"Limit DNS queries to a set of domain:[type,...] pairs" env:"DISCODNS_REJECT"`
                CPUProfile       bool     `long:"cpuprofile" description:"Enable CPU Profiling" env:"DISCODNS_CPUPROFILE"`
	}
)

func main() {

	_, err := flags.ParseArgs(&options, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

        if options.CPUProfile {
		f, err := os.Create("cpuprofile")
		if err != nil {
			log.Fatal(err)
			os.Exit(2)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
        }

	if options.Debug {
		logDebug = true
		debugMsg("Debug mode enabled")
	}

	// Create an ETCD client
	etcd := etcd.NewClient(options.EtcdHosts)
	if !etcd.SyncCluster() {
		logger.Printf("[WARNING] Failed to connect to etcd cluster at launch time")
	}

	// Register the metrics writer
	if len(options.GraphiteServer) > 0 {
		addr, err := net.ResolveTCPAddr("tcp", options.GraphiteServer)
		if err != nil {
			logger.Fatal("Failed to parse graphite server: ", err.Error())
		}

		prefix := "discodns"
		hostname, err := os.Hostname()
		if err != nil {
			logger.Fatal("Unable to get hostname: ", err.Error())
		}

		prefix = prefix + "." + strings.Replace(hostname, ".", "_", -1)

		go metrics.Graphite(metrics.DefaultRegistry, time.Duration(options.GraphiteDuration)*time.Second, prefix, addr)
	} else if options.MetricsDuration > 0 {
		go metrics.Log(metrics.DefaultRegistry, time.Duration(options.MetricsDuration)*time.Second, logger)

		// Register a bunch of debug metrics
		metrics.RegisterDebugGCStats(metrics.DefaultRegistry)
		metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)
		go metrics.CaptureDebugGCStats(metrics.DefaultRegistry, time.Duration(options.MetricsDuration))
		go metrics.CaptureRuntimeMemStats(metrics.DefaultRegistry, time.Duration(options.MetricsDuration))
	} else {
		logger.Printf("Metric logging disabled")
	}

	// Start up the DNS resolver server
	server := &server{
		addr:       options.ListenAddress,
		port:       options.ListenPort,
		etcd:       etcd,
		rTimeout:   time.Duration(5) * time.Second,
		wTimeout:   time.Duration(5) * time.Second,
		defaultTTL: options.DefaultTTL,
		queryFilterer: &QueryFilterer{
			acceptFilters: parseFilters(options.Accept),
			rejectFilters: parseFilters(options.Reject)},
	}

	server.Run()

	logger.Printf("Listening on %s:%d\n", options.ListenAddress, options.ListenPort)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			logger.Printf("Bye bye :(\n")
			break forever
		}
	}
}

func debugMsg(v ...interface{}) {
	if logDebug {
		vars := []interface{}{"[", runtime.NumGoroutine(), "]"}
		vars = append(vars, v...)

		logger.Println(vars...)
	}
}

// parseFilters will convert a string into a Query Filter structure. The accepted
// format for input is [domain]:[type,type,...]. For example...
//
// - "domain:A,AAAA" # Match all A and AAAA queries within `domain`
// - ":TXT" # Matches only TXT queries for any domain
// - "domain:" # Matches any query within `domain`
func parseFilters(filters []string) []QueryFilter {
	var parsedFilters []QueryFilter
	for _, filter := range filters {
		components := strings.Split(filter, ":")
		if len(components) != 2 {
			logger.Printf("Expected only one colon ([domain]:[type,type...])")
			continue
		}

		domain := dns.Fqdn(components[0])
		types := strings.Split(components[1], ",")

		if len(types) == 1 && len(types[0]) == 0 {
			types = make([]string, 0)
		}

		debugMsg("Adding filter with domain '" + domain + "' and types '" + strings.Join(types, ",") + "'")
		parsedFilters = append(parsedFilters, QueryFilter{domain, types})
	}

	return parsedFilters
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
