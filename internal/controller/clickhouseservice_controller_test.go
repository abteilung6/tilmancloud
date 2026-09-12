package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clickhousev1alpha1 "github.com/abteilung6/tilmancloud/api/v1alpha1"
)

// sameRequest is the input the manager passes into Reconcile: namespace + name only.
func sameRequest() ctrl.Request {
	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "clickhouse-managed",
			Name:      "clickhouse",
		},
	}
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clickhousev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
}

func TestReconcileExistingClickHouseService(t *testing.T) {
	scheme := testScheme(t)
	existing := &clickhousev1alpha1.ClickHouseService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clickhouse",
			Namespace: "clickhouse-managed",
		},
	}
	r := &ClickHouseServiceReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build(),
		Scheme: scheme,
	}

	result, err := r.Reconcile(context.Background(), sameRequest())
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if !result.IsZero() {
		t.Fatalf("expected no requeue, got %+v", result)
	}
}

func TestReconcileMissingClickHouseService(t *testing.T) {
	scheme := testScheme(t)
	r := &ClickHouseServiceReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
		Scheme: scheme,
	}

	result, err := r.Reconcile(context.Background(), sameRequest())
	if err != nil {
		t.Fatalf("deleted object must not be an error: %v", err)
	}
	if !result.IsZero() {
		t.Fatalf("expected no requeue, got %+v", result)
	}
}
