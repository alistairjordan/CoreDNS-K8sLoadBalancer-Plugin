package loadbalancer

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const ttl = 30

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("k8s_loadbalancer")

// LoadBalancer is a plugin to show how to write a plugin.
type LoadBalancer struct {
	KubeConfigPath string
	EnableRootZone bool
	EnableNSZone   bool
	RootZones      []string
	RecordsSync    *sync.Mutex
	Records        []kubeRecord
	Next           plugin.Handler
}

func (e LoadBalancer) RemoveZoneSuffix(name string, suffix []string) string {
	for _, s := range suffix {
		if strings.HasSuffix(name, s) {
			length := len([]rune(name)) - len([]rune(s))
			subname := name[:length]
			return subname
		}
	}
	return ""
}

// ServeDNS implements the plugin.Handler interface. This method gets called when kube zone is used
// in a Server.
func (e LoadBalancer) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.QName()
	log.Debug("Received response")
	log.Debugf("Intercepting localhost query for %q %s, from %s", state.Name(), state.Type(), state.IP())
	pw := NewResponsePrinter(w)
	name := e.RemoveZoneSuffix(state.Name(), e.RootZones)
	// Check if in .kube subdomain
	if name != "" {
		log.Debugf("Finding record %s", name)
		record, found := e.getRecord(name)
		if found {
			log.Debugf("Found Record %+v", record)

			// Export metric with the server label set to the current server handling the request.
			requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

			m := new(dns.Msg)
			m.SetReply(state.Req)
			//var records []dns.RR
			hdr := dns.RR_Header{Name: qname, Ttl: ttl, Class: dns.ClassINET, Rrtype: dns.TypeA}
			m.Answer = []dns.RR{&dns.A{Hdr: hdr, A: net.ParseIP(record.ip).To4()}}
			//m.Ns = soaFromOrigin(qname)
			//m.Answer = records
			log.Debugf("Responding with %+v", m.Answer)
			w.WriteMsg(m)
		}

	}
	// Call next plugin (if any).
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, pw, r)
}

// Name implements the Handler interface.
func (e LoadBalancer) Name() string { return "k8s_loadbalancer" }

func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{}
	lb.RecordsSync = &sync.Mutex{}
	lb.EnableNSZone = true
	lb.EnableRootZone = true
	lb.RootZones = []string{".kube."}
	return lb
}

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("LoadBalancer")
	return r.ResponseWriter.WriteMsg(res)
}
