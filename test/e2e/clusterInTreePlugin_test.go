// +build intree

package e2e

import (
	goctx "context"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	storageos "github.com/storageos/cluster-operator/pkg/apis/storageos/v1"
	testutil "github.com/storageos/cluster-operator/test/e2e/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestClusterInTreePlugin(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	resourceNS := "storageos"

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	clusterSpec := storageos.StorageOSClusterSpec{
		SecretRefName:      "storageos-api",
		SecretRefNamespace: "default",
		ResourceNS:         resourceNS,
		Tolerations: []corev1.Toleration{
			{
				Key:      "key",
				Operator: corev1.TolerationOpEqual,
				Value:    "value",
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
		K8sDistro: "openshift",
	}

	testStorageOS := testutil.NewStorageOSCluster(namespace, clusterSpec)

	testutil.SetupOperator(t, ctx)
	err = testutil.DeployCluster(t, ctx, testStorageOS)
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-storageos", Namespace: namespace}, testStorageOS)
	if err != nil {
		t.Fatal(err)
	}

	testutil.ClusterStatusCheck(t, testStorageOS.Status, 1)

	daemonset, err := f.KubeClient.AppsV1().DaemonSets(resourceNS).Get("storageos-daemonset", metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		t.Fatalf("failed to get storageos-daemonset: %v", err)
	}

	// Check the number of containers in daemonset pod spec.
	if len(daemonset.Spec.Template.Spec.Containers) != 1 {
		t.Errorf("unexpected number of daemonset pod containers:\n\t(GOT) %d\n\t(WNT) %d", len(daemonset.Spec.Template.Spec.Containers), 2)
	}

	testutil.NodeLabelSyncTest(t, f.KubeClient)
}
