package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clickhousev1alpha1 "github.com/abteilung6/tilmancloud/api/v1alpha1"
)

// ClickHouseServiceReconciler watches ClickHouseService and applies the
// lab-shaped ConfigMap, Secret, Services, and StatefulSet.
type ClickHouseServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=clickhouse.tilmancloud.io,resources=clickhouseservices,verbs=get;list;watch
// +kubebuilder:rbac:groups=clickhouse.tilmancloud.io,resources=clickhouseservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services;configmaps;secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch

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

	if obj.Namespace == labNamespace {
		logger.Info("refusing to reconcile in the lab namespace", "namespace", obj.Namespace)
		return ctrl.Result{}, nil
	}

	children := []client.Object{
		desiredConfigMap(&obj),
		desiredSecret(&obj),
		desiredService(&obj),
		desiredHeadlessService(&obj),
		desiredStatefulSet(&obj),
	}
	for _, child := range children {
		if err := ctrl.SetControllerReference(&obj, child, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Patch(ctx, child, client.Apply, client.ForceOwnership, client.FieldOwner(fieldOwner)); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.syncStatus(ctx, &obj); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("applied lab-shaped children", "namespace", req.Namespace, "name", req.Name)
	return ctrl.Result{}, nil
}

func (r *ClickHouseServiceReconciler) syncStatus(ctx context.Context, obj *clickhousev1alpha1.ClickHouseService) error {
	cond := metav1.Condition{
		Type:               "Ready",
		ObservedGeneration: obj.Generation,
	}
	ready := int32(0)

	var sts appsv1.StatefulSet
	err := r.Get(ctx, client.ObjectKeyFromObject(obj), &sts)
	switch {
	case apierrors.IsNotFound(err):
		cond.Status = metav1.ConditionFalse
		cond.Reason = "StatefulSetNotFound"
	case err != nil:
		return err
	case sts.Status.ReadyReplicas == 1:
		ready = 1
		cond.Status = metav1.ConditionTrue
		cond.Reason = "StatefulSetReady"
	default:
		ready = sts.Status.ReadyReplicas
		cond.Status = metav1.ConditionFalse
		cond.Reason = "StatefulSetNotReady"
	}

	orig := obj.DeepCopy()
	obj.Status.ReadyReplicas = ready
	meta.SetStatusCondition(&obj.Status.Conditions, cond)
	return r.Status().Patch(ctx, obj, client.MergeFrom(orig))
}

func (r *ClickHouseServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clickhousev1alpha1.ClickHouseService{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Named("clickhouseservice").
		Complete(r)
}
