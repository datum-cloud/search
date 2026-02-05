package v1alpha1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"go.miloapis.net/search/internal/cel"
	"go.miloapis.net/search/internal/jsonpath"
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

	var allErrs field.ErrorList

	// Validate CEL expressions in conditions
	for i, condition := range obj.Spec.Conditions {
		errs := v.CelValidator.Validate(condition.Expression)
		for _, err := range errs {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec", "conditions").Index(i).Child("expression"),
				condition.Expression,
				err,
			))
		}
	}

	// Validate JSONPath in fields
	for i, fieldPolicy := range obj.Spec.Fields {
		if err := jsonpath.ValidatePath(fieldPolicy.Path); err != "" {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec", "fields").Index(i).Child("path"),
				fieldPolicy.Path,
				err,
			))
		}
	}

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
