package loadbalancer

import (
	"bytes"
	"context"
	golog "log"
	"strings"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func TestRemoveZoneSuffix(t *testing.T) {
	x := LoadBalancer{Next: test.ErrorHandler()}
	tests := []struct {
		zones  []string
		query  string
		result string
	}{
		{
			zones:  []string{".kube.", "kubernetes"},
			query:  "testabc.kube.",
			result: "testabc",
		},
	}
	for _, test := range tests {
		result := x.RemoveZoneSuffix(test.query, test.zones)
		if result != test.result {
			t.Errorf("Queried %s against %+v expecting %s, but got %s", test.query, test.zones, test.result, result)
		}
	}
}
func TestLoadBalancer(t *testing.T) {
	// Create a new LoadBalancer Plugin. Use the test.ErrorHandler as the next plugin.
	x := LoadBalancer{Next: test.ErrorHandler()}

	// Setup a new output buffer that is *not* standard output, so we can check if
	// LoadBalancer is really being printed.
	b := &bytes.Buffer{}
	golog.SetOutput(b)

	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion("example.org.", dns.TypeA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	// Call our plugin directly, and check the result.
	x.ServeDNS(ctx, rec, r)
	if a := b.String(); !strings.Contains(a, "[INFO] plugin/k8s_loadbalancer: LoadBalancer") {
		t.Errorf("Failed to print '%s', got %s", "[INFO] plugin/k8s_loadbalancer: LoadBalancer", a)
	}
}
