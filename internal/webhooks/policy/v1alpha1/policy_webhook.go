package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"go.miloapis.net/search/internal/cel"
	"go.miloapis.net/search/internal/policy/validation"
	policyv1alpha1 "go.miloapis.net/search/pkg/apis/policy/v1alpha1"
)

var resourceIndexPolicies = policyv1alpha1.Resource("resourceindexpolicies")

// +kubebuilder:webhook:path=/validate-policy-search-miloapis-com-v1alpha1-resourceindexpolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=policy.search.miloapis.com,resources=resourceindexpolicies,verbs=create;update,versions=v1alpha1,name=vresourceindexpolicy.kb.io,admissionReviewVersions=v1

// ResourceIndexPolicyValidator validates ResourceIndexPolicy resources.
type ResourceIndexPolicyValidator struct {
	CelValidator *cel.Validator
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ResourceIndexPolicyValidator) ValidateCreate(ctx context.Context, obj *policyv1alpha1.ResourceIndexPolicy) (admission.Warnings, error) {
	logger := log.FromContext(ctx)

	logger.Info("Validating ResourceIndexPolicy")

	allErrs := validation.ValidateResourceIndexPolicy(obj, v.CelValidator)

	if len(allErrs) > 0 {
		return nil, errors.NewInvalid(policyv1alpha1.SchemeGroupVersion.WithKind("ResourceIndexPolicy").GroupKind(), obj.Name, allErrs)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ResourceIndexPolicyValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *policyv1alpha1.ResourceIndexPolicy) (admission.Warnings, error) {
	warningMsg := "ResourceIndexPolicy updates are not supported. Consider deleting the resource and creating a new one with the desired spec"
	return admission.Warnings{warningMsg}, errors.NewMethodNotSupported(resourceIndexPolicies, "update")
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ResourceIndexPolicyValidator) ValidateDelete(ctx context.Context, obj *policyv1alpha1.ResourceIndexPolicy) (admission.Warnings, error) {
	return nil, nil
}
