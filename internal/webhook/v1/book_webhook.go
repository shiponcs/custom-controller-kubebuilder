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

package v1

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	storev1 "github.com/shiponcs/custom-controller-kubebuilder/api/v1"
)

// nolint:unused
// log is for logging in this package.
var booklog = logf.Log.WithName("book-resource")

// SetupBookWebhookWithManager registers the webhook for Book in the manager.
func SetupBookWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&storev1.Book{}).
		WithValidator(&BookCustomValidator{}).
		WithDefaulter(&BookCustomDefaulter{
			Replicas:      1,
			ContainerPort: 8080,
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-store-crd-com-v1-book,mutating=true,failurePolicy=fail,sideEffects=None,groups=store.crd.com,resources=books,verbs=create;update,versions=v1,name=mbook-v1.kb.io,admissionReviewVersions=v1

// BookCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Book when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type BookCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
	Replicas      int32
	ContainerPort int32
}

var _ webhook.CustomDefaulter = &BookCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Book.
func (d *BookCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	book, ok := obj.(*storev1.Book)

	if !ok {
		return fmt.Errorf("expected an Book object but got %T", obj)
	}
	booklog.Info("Defaulting for Book 1", "name", book.GetName())

	// TODO(user): fill in your defaulting logic.
	if book.Spec.Replicas == nil || *book.Spec.Replicas == 0 {
		book.Spec.Replicas = new(int32)
		*book.Spec.Replicas = d.Replicas
	}

	if len(book.Spec.Container.Ports) == 0 {
		booklog.Info("Defaulting for Book container ports is empty", "name", book.GetName())
		book.Spec.Container.Ports = []corev1.ContainerPort{
			{ContainerPort: d.ContainerPort},
		}
	} else if book.Spec.Container.Ports[0].ContainerPort == 0 {
		booklog.Info("Defaulting for Book container port is empty", "name", book.GetName())
		book.Spec.Container.Ports = []corev1.ContainerPort{
			{ContainerPort: d.ContainerPort},
		}
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-store-crd-com-v1-book,mutating=false,failurePolicy=fail,sideEffects=None,groups=store.crd.com,resources=books,verbs=create;update,versions=v1,name=vbook-v1.kb.io,admissionReviewVersions=v1

// BookCustomValidator struct is responsible for validating the Book resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type BookCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &BookCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Book.
func (v *BookCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	book, ok := obj.(*storev1.Book)
	if !ok {
		return nil, fmt.Errorf("expected a Book object but got %T", obj)
	}
	booklog.Info("Validation for Book upon creation", "name", book.GetName())

	// TODO(user): fill in your validation logic upon object creation.
	if book.Spec.Container.Image == "" {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "store.crd.com", Kind: "Book"}, book.Name, field.ErrorList{field.Invalid(field.NewPath(book.ObjectMeta.Name).Child("container"), "container", "must have image value set")},
		)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Book.
func (v *BookCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	book, ok := newObj.(*storev1.Book)
	if !ok {
		return nil, fmt.Errorf("expected a Book object for the newObj but got %T", newObj)
	}
	booklog.Info("Validation for Book upon update", "name", book.GetName())

	if book.Spec.Container.Image == "" {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "store.crd.com", Kind: "Book"}, book.Name, field.ErrorList{field.Invalid(field.NewPath(book.ObjectMeta.Name).Child("container"), "container", "must have image value set")},
		)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Book.
func (v *BookCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	book, ok := obj.(*storev1.Book)
	if !ok {
		return nil, fmt.Errorf("expected a Book object but got %T", obj)
	}
	booklog.Info("Validation for Book upon deletion", "name", book.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
