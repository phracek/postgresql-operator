package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	postgresqlv1 "github.com/mcyprian/postgresql-operator/pkg/apis/postgresql/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PostgreSQLClusterScaling test downscaling of slave node and rejoining the cluster again
func PostgreSQLClusterScaling(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	// get global framework variables reference
	f := framework.Global

	initializeTestEnvironment(t, f, ctx)

	if err := postgreSQLClusterScalingTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func postgreSQLClusterScalingTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Couldn't get namespace: %v", err)
	}
	exampleName := types.NamespacedName{Name: postgreSQLCRName, Namespace: namespace}
	examplePostgreSQL := newTestCluster(namespace)

	if err := f.Client.Create(goctx.TODO(), examplePostgreSQL, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval}); err != nil {
		return fmt.Errorf("Failed to create example PostgreSQL: %v", err)
	}
	if err := e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "primary-node", 1, retryInterval, timeout); err != nil {
		return fmt.Errorf("Waiting for deployment primary-node timed out: %v", err)
	}
	if err := e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "standby-node", 1, retryInterval, timeout); err != nil {
		return fmt.Errorf("Waiting for deployment standby-node timed out: %v", err)
	}
	if err := retryExecution(t, f, namespace, getStatusDouble, 7, time.Second*10); err != nil {
		return err
	}
	t.Log("Initial deployment created.")
	standbyDeployment, err := f.KubeClient.AppsV1().Deployments(namespace).Get("standby-node", metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return fmt.Errorf("Failed to get standby-node deployment: %v", err)
	}
	current := &postgresqlv1.PostgreSQL{}
	if err := f.Client.Get(goctx.TODO(), exampleName, current); err != nil {
		return fmt.Errorf("Failed to get examplePostgreSQL: %v", err)
	}
	delete(current.Spec.Nodes, "standby-node")

	if err := f.Client.Update(goctx.TODO(), current); err != nil {
		return fmt.Errorf("Failed to update cluster: %v", err)
	}
	if err := e2eutil.WaitForDeletion(t, f.Client.Client, standbyDeployment, retryInterval, timeout); err != nil {
		return fmt.Errorf("Waiting for standby-node deletion timed out: %v", err)
	}
	if err = retryExecution(t, f, namespace, getStatusSingle, 7, time.Second*10); err != nil {
		return err
	}
	t.Log("Downscale success.")
	if err := f.Client.Get(goctx.TODO(), exampleName, current); err != nil {
		return fmt.Errorf("Failed to get examplePostgreSQL: %v", err)
	}
	current.Spec.Nodes["standby-node"] = examplePostgreSQL.Spec.Nodes["standby-node"]

	if err := f.Client.Update(goctx.TODO(), current); err != nil {
		return fmt.Errorf("Failed to update cluster: %v", err)
	}

	if err := e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "standby-node", 1, retryInterval, timeout); err != nil {
		return fmt.Errorf("Waiting for deployment standby-node timed out: %v", err)
	}
	if err = retryExecution(t, f, namespace, getStatusDouble, 7, time.Second*10); err != nil {
		return err
	}
	t.Log("Upscale success.")

	t.Log("Success")
	return nil
}
