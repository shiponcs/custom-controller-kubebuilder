/*
Copyright 2025.

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

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	storev1 "github.com/shiponcs/custom-controller-kubebuilder/api/v1"
)

// BookReconciler reconciles a Book object
type BookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	deploymentOwnderKey = ".metadata.controller"
	svcOwnerKey         = ".metadata.controller"
)

// +kubebuilder:rbac:groups=store.crd.com,resources=books,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=store.crd.com,resources=books/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=store.crd.com,resources=books/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Book object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Book")

	var bookCR storev1.Book
	if err := r.Get(ctx, req.NamespacedName, &bookCR); err != nil {
		logger.Error(err, "unable to fetch Book")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var childDeploys appsv1.DeploymentList
	if err := r.List(ctx, &childDeploys, client.InNamespace(req.Namespace), client.MatchingFields{deploymentOwnderKey: req.Name}); err != nil {
		logger.Error(err, "unable to list child Deployments")
		return ctrl.Result{}, err
	}

	if len(childDeploys.Items) > 0 {
		bookCR.Status.AvailableReplicas = childDeploys.Items[0].Status.AvailableReplicas
	} else {
		bookCR.Status.AvailableReplicas = 0
	}

	if err := r.Status().Update(ctx, &bookCR); err != nil {
		logger.Error(err, "unable to update Book")
		return ctrl.Result{}, err
	}

	if len(childDeploys.Items) == 0 {
		deploy := newDeployment(&bookCR)
		err := r.Create(ctx, deploy)
		if err != nil {
			logger.Error(err, "unable to create child Deployment ", deploy)
			return ctrl.Result{}, err
		}
	}

	if bookCR.Spec.Replicas != nil && len(childDeploys.Items) != 0 && *bookCR.Spec.Replicas != *childDeploys.Items[0].Spec.Replicas {
		deploy := newDeployment(&bookCR)
		err := r.Update(ctx, deploy)
		if err != nil {
			logger.Error(err, "unable to update child Deployment ", deploy)
			return ctrl.Result{}, err
		}
	}

	var childSvcs corev1.ServiceList
	if err := r.List(ctx, &childSvcs, client.InNamespace(req.Namespace), client.MatchingFields{svcOwnerKey: req.Name}); err != nil {
		logger.Error(err, "unable to list child services")
		return ctrl.Result{}, err
	}

	if len(childSvcs.Items) == 0 {
		svc := newService(&bookCR)
		err := r.Create(ctx, svc)
		if err != nil {
			logger.Error(err, "unable to create child Service")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// newDeployment creates a new Deployment for a book resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the book resource that 'owns' it.
func newDeployment(book *storev1.Book) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "book-server",
		"controller": book.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      book.Spec.DeploymentName + "-dep",
			Namespace: book.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(book, storev1.GroupVersion.WithKind("Book")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: book.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						book.Spec.Container,
					},
				},
			},
		},
	}
}

func newService(book *storev1.Book) *corev1.Service {
	labels := map[string]string{
		"app":        "book-server",
		"controller": book.Name,
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      book.Spec.DeploymentName + "service",
			Namespace: book.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(book, storev1.GroupVersion.WithKind("Book")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       book.Spec.Container.Ports[0].ContainerPort,
					TargetPort: intstr.FromInt32(book.Spec.Container.Ports[0].ContainerPort),
					NodePort:   30009,
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *BookReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, deploymentOwnderKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		deploy := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deploy)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != storev1.GroupVersion.String() || owner.Kind != "Book" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, svcOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		svc := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(svc)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != storev1.GroupVersion.String() || owner.Kind != "Book" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&storev1.Book{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("book").
		Complete(r)
}
