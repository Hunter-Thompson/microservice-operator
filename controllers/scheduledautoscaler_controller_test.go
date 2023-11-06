package controllers

import (
	"context"
	"os"
	"testing"
	"time"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	v2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func setupsaTestDeps(t *testing.T) (logr.Logger, *ScheduledAutoscalerReconciler) {
	s := prepareSchema(t, scheme.Scheme)
	r := ScheduledAutoscalerReconciler{
		Client:    k8sClient,
		Scheme:    s,
		Resources: resources.NewResourceHelper(k8sClient, s),
		allcron:   cron.New(),
	}

	logger := log.FromContext(context.TODO())
	logf.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))

	return logger, &r
}

func TestScheduledAutoscaler(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	_, r := setupsaTestDeps(t)
	r.allcron.Start()

	msName := "foo"
	msNamespace := "default"
	msImage := "image:latest"
	msLabels := map[string]string{
		"app": "test",
	}
	msReplicas := int32(8)

	ms := &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      msName,
			Namespace: msNamespace,
			UID:       types.UID("test"),
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    msImage,
			Labels:   msLabels,
			Replicas: msReplicas,
		},
	}

	err := r.Create(context.TODO(), ms)
	assert.NoError(t, err)

	saName := "foo"
	saNamespace := "default"

	sa := &microservicev1beta1.ScheduledAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: saNamespace,
			UID:       types.UID("test"),
		},
	}

	t.Run("reconcile", func(t *testing.T) {
		hpa := &v2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      saName,
				Namespace: saNamespace,
				UID:       types.UID("test"),
			},
			Spec: v2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: v2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "foo",
				},
				MinReplicas: &msReplicas,
				MaxReplicas: msReplicas,
			},
		}

		err := r.Create(context.TODO(), hpa)
		assert.NoError(t, err)

		maxReplicas := int32(10)
		minReplicas := int32(1)
		sa.Spec = microservicev1beta1.ScheduledAutoscalerSpec{
			MicroserviceName: msName,
			Schedules: []microservicev1beta1.Schedule{
				{MinReplicas: minReplicas, MaxReplicas: maxReplicas, Cron: "@every 1s", Name: "test"},
			},
		}

		err = r.Client.Create(context.TODO(), sa)
		assert.NoError(t, err)

		req := reconcile.Request{NamespacedName: types.NamespacedName{Name: saName, Namespace: saNamespace}}
		_, err = r.Reconcile(context.TODO(), req)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(r.allcron.Entries()))
		time.Sleep(1 * time.Second)
		r.allcron.Stop()

		ms := &microservicev1beta1.Microservice{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, ms)
		assert.NoError(t, err)

		hpa = &v2.HorizontalPodAutoscaler{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, hpa)
		assert.NoError(t, err)

		assert.Equal(t, maxReplicas, hpa.Spec.MaxReplicas)
		assert.Equal(t, map[string]string{
			"scheduledautoscaler.override": "true",
		}, ms.Annotations)
	})
}
