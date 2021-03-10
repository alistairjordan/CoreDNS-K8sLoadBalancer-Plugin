package loadbalancer

import(
	"strings"
	"context"
    "os"
    "path/filepath"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	"sync"
)

type kubeRecord struct {
	namespace string
	name string
	ip string
}

func (e LoadBalancer) kubeConnect() (v1.CoreV1Interface) {
	// Check the mutex is has been initialised (saves a load of effort with the tests)
	if e.RecordsSync == nil {
		e.RecordsSync = &sync.Mutex{}
	}
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
   )
   config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
   if err != nil {
	   log.Fatal(err)
   }

   clientset, err := kubernetes.NewForConfig(config)
   if err != nil {
	   log.Fatal(err)
   }

   api := clientset.CoreV1()
   return api
}

func (e LoadBalancer) kubeRecords(api v1.CoreV1Interface) ([]kubeRecord, error) {
	var records []kubeRecord
	namespace := ""
	listOptions := metav1.ListOptions{

    }

    svcs, err := api.Services(namespace).List(context.TODO(),listOptions)
    if err != nil {
        log.Fatal(err)
		return records, err
    }

    for _, svc := range(svcs.Items) {
        if svc.Spec.Type == "LoadBalancer" {
            //fmt.Printf("%s\t%s\n",svc.ObjectMeta.Name,svc.Status.LoadBalancer.Ingress[0].IP)
			r := &kubeRecord{
				namespace: namespace,
				name: svc.ObjectMeta.Name,
				ip: svc.Status.LoadBalancer.Ingress[0].IP}
			records = append(records, *r)
        }
    }

	return records, err
}

func (e *LoadBalancer) updateRecords(records []kubeRecord) {
	//copy(e.Records,records)
	//fmt.Printf("%+v", e.Records)
	e.Records = nil
	for _, record := range(records) {
		e.RecordsSync.Lock()
		e.Records = append(e.Records, record)
		e.RecordsSync.Unlock()
		//log.Debugf("Appending %+v\n",record)
	}
}

func (e *LoadBalancer) getRecord(name string) (kubeRecord, bool) {
	e.RecordsSync.Lock()
	defer e.RecordsSync.Unlock()
	for _, record := range(e.Records) {
		log.Debugf("Comparing %s on record to %s",record.name, name)
		if strings.Contains(record.name, name) {
			return record, true
		}
	}
	return kubeRecord{}, false
}

func (e *LoadBalancer) updateTicker() {
	api := e.kubeConnect()
	for {
		records, err := e.kubeRecords(api)
		if err != nil {
			log.Fatalf("Record Update failed with %s", err)
		}
		e.updateRecords(records)
		log.Debugf("Updated %d records", len(records))
		time.Sleep(1600 * time.Millisecond)
	}
}