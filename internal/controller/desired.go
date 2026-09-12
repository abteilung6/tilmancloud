package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	clickhousev1alpha1 "github.com/abteilung6/tilmancloud/api/v1alpha1"
)

const (
	fieldOwner      = "clickhouse-service"
	clickhouseImage = "clickhouse/clickhouse-server:26.3.20.7@sha256:c22a4aba84a1ea83e320a6f8620b87100649e04995b61d7dde160fc1ca33bc76"
	listenXML       = `<clickhouse>
  <listen_host replace="replace">0.0.0.0</listen_host>
</clickhouse>
`
	labPassword  = "clickhouse-lab"
	labNamespace = "clickhouse-lab"
)

func labels(cr *clickhousev1alpha1.ClickHouseService) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "clickhouse",
		"app.kubernetes.io/instance":   cr.Name,
		"app.kubernetes.io/part-of":    "tilmancloud",
		"app.kubernetes.io/managed-by": "clickhouseservice",
	}
}

func selector(cr *clickhousev1alpha1.ClickHouseService) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "clickhouse",
		"app.kubernetes.io/instance": cr.Name,
	}
}

func objectMeta(cr *clickhousev1alpha1.ClickHouseService, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: cr.Namespace,
		Labels:    labels(cr),
	}
}

func ports() []corev1.ServicePort {
	return []corev1.ServicePort{
		{Name: "http", Port: 8123, TargetPort: intstr.FromString("http")},
		{Name: "native", Port: 9000, TargetPort: intstr.FromString("native")},
	}
}

func desiredConfigMap(cr *clickhousev1alpha1.ClickHouseService) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: objectMeta(cr, cr.Name),
		Data:       map[string]string{"listen.xml": listenXML},
	}
}

func desiredSecret(cr *clickhousev1alpha1.ClickHouseService) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: objectMeta(cr, cr.Name),
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"password": labPassword},
	}
}

func desiredService(cr *clickhousev1alpha1.ClickHouseService) *corev1.Service {
	return &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: objectMeta(cr, cr.Name),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selector(cr),
			Ports:    ports(),
		},
	}
}

func desiredHeadlessService(cr *clickhousev1alpha1.ClickHouseService) *corev1.Service {
	return &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: objectMeta(cr, cr.Name+"-headless"),
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector:  selector(cr),
			Ports:     ports(),
		},
	}
}

func pingProbe(initial, period, timeout, failure int32) corev1.Probe {
	return corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/ping",
				Port: intstr.FromString("http"),
			},
		},
		InitialDelaySeconds: initial,
		PeriodSeconds:       period,
		TimeoutSeconds:      timeout,
		FailureThreshold:    failure,
	}
}

func desiredStatefulSet(cr *clickhousev1alpha1.ClickHouseService) *appsv1.StatefulSet {
	ls := labels(cr)
	sel := selector(cr)
	readiness := pingProbe(10, 5, 5, 12)
	liveness := pingProbe(60, 10, 5, 6)
	return &appsv1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "StatefulSet"},
		ObjectMeta: objectMeta(cr, cr.Name),
		Spec: appsv1.StatefulSetSpec{
			Replicas:    ptr.To(int32(1)),
			ServiceName: cr.Name + "-headless",
			Selector:    &metav1.LabelSelector{MatchLabels: sel},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: ls},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot:   ptr.To(true),
						RunAsUser:      ptr.To(int64(101)),
						RunAsGroup:     ptr.To(int64(101)),
						FSGroup:        ptr.To(int64(101)),
						SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
					},
					Containers: []corev1.Container{{
						Name:            "clickhouse",
						Image:           clickhouseImage,
						ImagePullPolicy: corev1.PullIfNotPresent,
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
						},
						Ports: []corev1.ContainerPort{
							{Name: "http", ContainerPort: 8123},
							{Name: "native", ContainerPort: 9000},
						},
						Env: []corev1.EnvVar{{
							Name: "CLICKHOUSE_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name},
									Key:                  "password",
								},
							},
						}},
						ReadinessProbe: &readiness,
						LivenessProbe:  &liveness,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("500m"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "data", MountPath: "/var/lib/clickhouse"},
							{Name: "config", MountPath: "/etc/clickhouse-server/config.d"},
							{Name: "logs", MountPath: "/var/log/clickhouse-server"},
						},
					}},
					Volumes: []corev1.Volume{
						{Name: "config", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name}}}},
						{Name: "logs", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "PersistentVolumeClaim"},
				ObjectMeta: metav1.ObjectMeta{Name: "data"},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					StorageClassName: ptr.To("standard"),
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
					},
				},
			}},
		},
	}
}
