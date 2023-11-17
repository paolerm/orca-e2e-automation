/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	// appsv1 "k8s.io/api/apps/v1"
	"encoding/json"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	orcav1beta1 "github.com/paolerm/orca-e2e-automation/api/v1beta1"
)

// E2EAutomationReconciler reconciles a E2EAutomation object
type E2EAutomationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const e2eTestFinalizer = "paermini.com/e2e-test-finalizer"

//+kubebuilder:rbac:groups=orca.paermini.com,resources=e2eautomations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=orca.paermini.com,resources=e2eautomations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=orca.paermini.com,resources=e2eautomations/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the E2EAutomation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *E2EAutomationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Getting E2E test CR...")

	e2eTest := &orcav1beta1.E2EAutomation{}
	err := r.Get(ctx, req.NamespacedName, e2eTest)
	if err != nil {
		logger.Error(err, "Failed to get CR!")
		return ctrl.Result{}, err
	}

	isCrDeleted := e2eTest.GetDeletionTimestamp() != nil
	if isCrDeleted {
		if controllerutil.ContainsFinalizer(e2eTest, e2eTestFinalizer) {
			// Run finalization logic. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeE2eTest(ctx, req, e2eTest); err != nil {
				return ctrl.Result{}, err
			}

			// Remove e2eTestFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(e2eTest, e2eTestFinalizer)
			err := r.Update(ctx, e2eTest)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	settingsParam, err := json.Marshal(e2eTest.Spec.Settings)

	testContainer := corev1.Container{
		Name:  e2eTest.Spec.Id,
		Image: e2eTest.Spec.ImageId,
		Env: []corev1.EnvVar{
			{Name: "SETTINGS", Value: string(settingsParam)},
		},
	}

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e2eTest.Spec.Id,
			Namespace: req.NamespacedName.Namespace,
			Labels: map[string]string{
				"simulation": e2eTest.Spec.Id,
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          e2eTest.Spec.Schedule,
			ConcurrencyPolicy: "Replace",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers:    []corev1.Container{testContainer},
							RestartPolicy: "Never",
						},
					},
				},
			},
		},
	}

	logger.Info("Creating E2E test CR...")

	err = r.Create(ctx, cronJob)
	if err != nil {
		// TODO: ignore not found error
		logger.Error(err, "Failed to create CronJob!")
		return ctrl.Result{}, err
	}

	// TODO: update flow

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *E2EAutomationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orcav1beta1.E2EAutomation{}).
		Complete(r)
}

func (r *E2EAutomationReconciler) finalizeE2eTest(ctx context.Context, req ctrl.Request, e2eTest *orcav1beta1.E2EAutomation) error {
	logger := log.FromContext(ctx)

	cronJob := &batchv1.CronJob{}
	cronJobName := types.NamespacedName{
		Name:      e2eTest.Spec.Id,
		Namespace: req.NamespacedName.Namespace,
	}

	err := r.Get(ctx, cronJobName, cronJob)
	if err != nil {
		// TODO: ignore not found error
		logger.Error(err, "Failed to get CronJob!")
		return err
	}

	logger.Info("Removing CronJob under namespace " + req.NamespacedName.Namespace + " and name " + e2eTest.Spec.Id + "...")
	err = r.Delete(ctx, cronJob)
	if err != nil {
		logger.Error(err, "Failed to delete CronJob!")
		return err
	}

	logger.Info("Successfully finalized")
	return nil
}
