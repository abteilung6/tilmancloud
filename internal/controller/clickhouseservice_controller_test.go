package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clickhousev1alpha1 "github.com/abteilung6/tilmancloud/api/v1alpha1"
)

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
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
}

func sampleCR(namespace string) *clickhousev1alpha1.ClickHouseService {
	return &clickhousev1alpha1.ClickHouseService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clickhousev1alpha1.GroupVersion.String(),
			Kind:       "ClickHouseService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clickhouse",
			Namespace: namespace,
			UID:       "11111111-1111-1111-1111-111111111111",
		},
	}
}

func TestReconcileExistingClickHouseService(t *testing.T) {
	scheme := testScheme(t)
	r := &ClickHouseServiceReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleCR("clickhouse-managed")).Build(),
		Scheme: scheme,
	}

	result, err := r.Reconcile(context.Background(), sameRequest())
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if !result.IsZero() {
		t.Fatalf("expected no requeue, got %+v", result)
	}

	ctx := context.Background()
	ns := "clickhouse-managed"
	mustGet := func(obj client.Object, name string) {
		t.Helper()
		if err := r.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, obj); err != nil {
			t.Fatalf("get %T %s: %v", obj, name, err)
		}
	}

	var cm corev1.ConfigMap
	mustGet(&cm, "clickhouse")
	var secret corev1.Secret
	mustGet(&secret, "clickhouse")
	var svc corev1.Service
	mustGet(&svc, "clickhouse")
	var headless corev1.Service
	mustGet(&headless, "clickhouse-headless")
	var sts appsv1.StatefulSet
	mustGet(&sts, "clickhouse")

	if headless.Spec.ClusterIP != corev1.ClusterIPNone {
		t.Fatalf("headless ClusterIP want None, got %q", headless.Spec.ClusterIP)
	}
	if sts.Spec.ServiceName != "clickhouse-headless" {
		t.Fatalf("StatefulSet serviceName %s", sts.Spec.ServiceName)
	}
	if got := sts.Labels["app.kubernetes.io/instance"]; got != "clickhouse" {
		t.Fatalf("instance label %s", got)
	}
	if got := sts.Labels["app.kubernetes.io/instance"]; got == "lab" {
		t.Fatal("managed objects must not use instance=lab")
	}
	if len(sts.OwnerReferences) != 1 || sts.OwnerReferences[0].Kind != "ClickHouseService" {
		t.Fatalf("ownerReferences %+v", sts.OwnerReferences)
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

func TestReconcileRefusesLabNamespace(t *testing.T) {
	scheme := testScheme(t)
	r := &ClickHouseServiceReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleCR("clickhouse-lab")).Build(),
		Scheme: scheme,
	}

	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "clickhouse-lab", Name: "clickhouse"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	var sts appsv1.StatefulSet
	err := r.Get(context.Background(), types.NamespacedName{Namespace: "clickhouse-lab", Name: "clickhouse"}, &sts)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("lab namespace must not get a managed StatefulSet, err=%v", err)
	}
}
