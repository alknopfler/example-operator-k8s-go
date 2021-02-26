/*
Copyright 2021 alknopfler.

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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventfinderv1beta1 "github.com/alknopfler/example-operator-k8s-go/api/v1beta1"
)

// SagaFinderReconciler reconciles a SagaFinder object
type SagaFinderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=event-finder.example.org,resources=sagafinders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=event-finder.example.org,resources=sagafinders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cache.example.com,resources=sagafinders/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *SagaFinderReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sagafinder", req.NamespacedName)

	sagafinder := &eventfinderv1beta1.SagaFinder{}
	err := r.Get(ctx, req.NamespacedName, sagafinder)
	if err != nil{
		if errors.IsNotFound(err){
			log.Info("Sagafinder resource not found. Ignoring")
		}
		log.Info("SagaFinder not available")
		return ctrl.Result{},nil
	}

	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: sagafinder.Name, Namespace: sagafinder.Namespace}, found)

	if err != nil && errors.IsNotFound(err){
		dep := r.deploymentForSaga(sagafinder)
		log.Info("Creating new deployment for sagaFinder")
		err = r.Create(ctx,dep)

		return ctrl.Result{}, err
	}
	size := sagafinder.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}


	return ctrl.Result{}, nil
}

func (r *SagaFinderReconciler) deploymentForSaga( s *eventfinderv1beta1.SagaFinder) *appsv1.Deployment{

	ls := labelsForSaga(s.Name)
	replicas := s.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:"nginx:latest",
						Name:    "sagaFinder",
						Command: []string{},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "sagaFinder",
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(s, dep, r.Scheme)
	return dep
}

func labelsForSaga(name string) map[string]string {
	return map[string]string{"app": "sagaFinder", "sagaFinder_cr": name}
}


func (r *SagaFinderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventfinderv1beta1.SagaFinder{}).
		Complete(r)
}
