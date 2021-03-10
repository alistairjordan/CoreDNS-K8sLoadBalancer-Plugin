package loadbalancer

import (
	"testing"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/coredns/coredns/plugin/test"
)

func TestKubeConnect(t *testing.T) {
	x := NewLoadBalancer()
	x.Next = test.ErrorHandler()
	api := x.kubeConnect()
	listOptions := metav1.ListOptions{

    }

    svcs, err := api.Services("").List(context.TODO(),listOptions)
    if err != nil {
        t.Errorf("Unable to get services, error: %s", err)
    }
	_ = svcs
}

func TestKubeRecords(t *testing.T) {
	x := NewLoadBalancer()
	x.Next = test.ErrorHandler()
	api := x.kubeConnect()
	list, err := x.kubeRecords(api)
	if err != nil {
		t.Errorf("Unable to populate records: %s", err)
	}
	_ = list

}

func TestUpdateRecords(t *testing.T) {
	x := NewLoadBalancer()
	x.Next = test.ErrorHandler()
	records := []kubeRecord{
		{
			namespace: "",
			name: "test1",
			ip: "10.1.1.1",
		},
		{
			namespace: "a",
			name: "test2",
			ip: "10.1.1.2",
		},
	}
	x.updateRecords(records)
	if x.Records[0].name != "test1" && x.Records[1].name != "test2" {
		t.Errorf("Expected\n %+v\n\n, got\n %+v\n\n", records, x.Records)
	}
}

func TestGetRecords(t *testing.T) {
	x := NewLoadBalancer()
	x.Next = test.ErrorHandler()
	records := []kubeRecord{
		{
			namespace: "",
			name: "test1.kube.",
			ip: "10.1.1.1",
		},
		{
			namespace: "a",
			name: "test2",
			ip: "10.1.1.2",
		},
	}
	x.updateRecords(records)
	record, found := x.getRecord("test2")
	if !found {
		t.Error("Record not found")
	}
	_ = record
}