package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 120
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

/*
CreateClusterLoggingCR this function is to create ClusterLogging instance in the clusterlogging namespace
*/
func CreateClusterLoggingCR(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, collector string, ESnodecount int32, storageclass string) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Could not get namespace: %v", err)
	}

	var collectionSpec logging.CollectionSpec
	if collector == "fluentd" {
		collectionSpec = logging.CollectionSpec{
			Logs: logging.LogCollectionSpec{
				Type:        logging.LogCollectionTypeFluentd,
				FluentdSpec: logging.FluentdSpec{},
			},
		}
	}
	if collector == "rsyslog" {
		collectionSpec = logging.CollectionSpec{
			Logs: logging.LogCollectionSpec{
				Type:        logging.LogCollectionTypeRsyslog,
				RsyslogSpec: logging.RsyslogSpec{},
			},
		}
	}

	storageSize := resource.NewQuantity(5*1024*1024*1024, resource.BinarySI)
	cpuValue, _ := resource.ParseQuantity("500m")
	memValue, _ := resource.ParseQuantity("2Gi")
	// create clusterlogging custom resource
	exampleClusterLogging := &logging.ClusterLogging{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterLogging",
			APIVersion: logging.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: namespace,
		},
		Spec: logging.ClusterLoggingSpec{
			LogStore: logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
				ElasticsearchSpec: logging.ElasticsearchSpec{
					NodeCount:        ESnodecount,
					RedundancyPolicy: "SingleRedundancy",
					Storage: v1alpha1.ElasticsearchStorageSpec{
						StorageClassName: storageclass,
						Size:             storageSize,
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    cpuValue,
							v1.ResourceMemory: memValue,
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    cpuValue,
							v1.ResourceMemory: memValue,
						},
					},
				},
			},
			Visualization: logging.VisualizationSpec{
				Type: logging.VisualizationTypeKibana,
				KibanaSpec: logging.KibanaSpec{
					Replicas: 1,
				},
			},
			Curation: logging.CurationSpec{
				Type: logging.CurationTypeCurator,
				CuratorSpec: logging.CuratorSpec{
					Schedule: "*/10 * * * *",
				},
			},
			Collection:      collectionSpec,
			ManagementState: logging.ManagementStateManaged,
		},
	}
	err = f.Client.Create(goctx.TODO(), exampleClusterLogging, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "kibana", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForCronJob(t, f.KubeClient, namespace, "curator", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = WaitForDaemonSet(t, f.KubeClient, namespace, collector, retryInterval, timeout)
	if err != nil {
		return err
	}

	return nil
}

/*
GetPodList to get pod match selector in a namespace
*/
func GetPodList(namespace string, selector string) (*core.PodList, error) {
	list := &core.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: core.SchemeGroupVersion.String(),
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}

/*
GetPodName to check indices in ES pod
*/
func GetPodName(t *testing.T, client framework.FrameworkClient, namespace string, container string, retryInterval, timeout time.Duration) (string, error) {
	pod := &v1.Pod{}
	podname := ""
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = client.Get(goctx.Background(), dynclient.ObjectKey{
			Namespace: namespace,
		}, pod)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of pod\n")
				return false, err
			}
			return false, err
		}
		if pod.Spec.Containers[0].Name == container {
			//			podname := pod.ObjectMeta.Name
			podname := pod.Name
			t.Logf("Find %s pod %s \n", container, podname)
			return true, nil
		}
		t.Logf("Couldn't find %s pod\n", container)
		return false, err
	})

	if err != nil {
		return podname, err
	}
	return podname, nil
}
