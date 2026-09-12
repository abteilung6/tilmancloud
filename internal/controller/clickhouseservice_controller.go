package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clickhousev1alpha1 "github.com/abteilung6/tilmancloud/api/v1alpha1"
)

// ClickHouseServiceReconciler watches ClickHouseService.
// It does not create a StatefulSet yet.
type ClickHouseServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=clickhouse.tilmancloud.io,resources=clickhouseservices,verbs=get;list;watch
// +kubebuilder:rbac:groups=clickhouse.tilmancloud.io,resources=clickhouseservices/status,verbs=get;update;patch

// Reconcile is called with a namespace+name key after a ClickHouseService
// is created, updated, or deleted. Read the object again; do not trust a diff.
func (r *ClickHouseServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var obj clickhousev1alpha1.ClickHouseService
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("reconcile ClickHouseService", "namespace", req.Namespace, "name", req.Name)
	return ctrl.Result{}, nil
}

func (r *ClickHouseServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clickhousev1alpha1.ClickHouseService{}).
		Named("clickhouseservice").
		Complete(r)
}
