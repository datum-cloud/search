package v1alpha1

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.miloapis.net/search/internal/cel"
	policyv1alpha1 "go.miloapis.net/search/pkg/apis/policy/v1alpha1"
)

func TestValidateCreate(t *testing.T) {
	// Create a CEL validator with reasonable depth
	celValidator, err := cel.NewValidator(10)
	if err != nil {
		t.Fatalf("failed to create CEL validator: %v", err)
	}

	validator := &ResourceIndexPolicyValidator{
		CelValidator: celValidator,
	}

	tests := []struct {
		name        string
		policy      *policyv1alpha1.ResourceIndexPolicy
		wantErr     bool
		errContains string
	}{
		{
			name: "valid policy with simple equality",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-1",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "is-active",
							Expression: "status.active == true",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid policy with multiple conditions",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-2",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "is-active",
							Expression: "status.phase == 'Active'",
						},
						{
							Name:       "has-label",
							Expression: "has(metadata.labels.app)",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid policy with logical operators",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-3",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "complex-condition",
							Expression: "status.ready == true && spec.enabled == true",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid policy with string functions",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-4",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "name-starts-with",
							Expression: "metadata.name.startsWith('prod-')",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid policy with comparison operators",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-5",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "count-check",
							Expression: "spec.replicas >= 1",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid policy with or operator",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-6",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "or-condition",
							Expression: "status.phase == 'Active' || status.phase == 'Running'",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid policy with ternary operator",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-7",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "ternary-condition",
							Expression: "has(spec.enabled) ? spec.enabled == true : false",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid policy - syntax error",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-invalid-1",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "bad-syntax",
							Expression: "status.active ==", // Missing right operand
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid policy - non-boolean expression",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-invalid-2",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "returns-string",
							Expression: "metadata.name", // Returns string, not boolean
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr:     true,
			errContains: "must evaluate to a boolean",
		},
		{
			name: "invalid policy - disallowed operator (arithmetic)",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-invalid-3",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "uses-arithmetic",
							Expression: "spec.count + 1 > 5", // Arithmetic + not allowed
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name: "invalid policy - disallowed function (type)",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-invalid-4",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "uses-type",
							Expression: "type(spec.count) == int", // type() function not allowed
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name: "invalid policy - one valid and one invalid condition",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-invalid-5",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "valid-condition",
							Expression: "status.active == true",
						},
						{
							Name:       "invalid-condition",
							Expression: "spec.count + 1 > 5", // This one is invalid
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid policy - undeclared variable",
			policy: &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy-invalid-6",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "uses-object",
							Expression: "object.status.active == true", // 'object' is not declared
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       ".spec.name",
							Searchable: true,
						},
					},
				},
			},
			wantErr:     true,
			errContains: "undeclared reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := validator.ValidateCreate(context.Background(), tt.policy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCreate() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateCreate() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCreate() unexpected error: %v", err)
				}
			}

			// We don't expect warnings on create
			if len(warnings) > 0 {
				t.Logf("ValidateCreate() warnings: %v", warnings)
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	celValidator, err := cel.NewValidator(10)
	if err != nil {
		t.Fatalf("failed to create CEL validator: %v", err)
	}

	validator := &ResourceIndexPolicyValidator{
		CelValidator: celValidator,
	}

	oldPolicy := &policyv1alpha1.ResourceIndexPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
		Spec: policyv1alpha1.ResourceIndexPolicySpec{
			TargetResource: policyv1alpha1.TargetResource{
				Group:   "contacts.miloapis.com",
				Version: "v1",
				Kind:    "Contact",
			},
			Conditions: []policyv1alpha1.PolicyCondition{
				{
					Name:       "is-active",
					Expression: "status.active == true",
				},
			},
			Fields: []policyv1alpha1.FieldPolicy{
				{
					Path:       ".spec.name",
					Searchable: true,
				},
			},
		},
	}

	newPolicy := oldPolicy.DeepCopy()
	newPolicy.Spec.Conditions[0].Expression = "status.phase == 'Active'"

	warnings, err := validator.ValidateUpdate(context.Background(), oldPolicy, newPolicy)

	// Update should always fail
	if err == nil {
		t.Error("ValidateUpdate() expected error but got none")
	}

	// Should have a warning about updates not being supported
	if len(warnings) == 0 {
		t.Error("ValidateUpdate() expected warning but got none")
	}
}

func TestValidateDelete(t *testing.T) {
	celValidator, err := cel.NewValidator(10)
	if err != nil {
		t.Fatalf("failed to create CEL validator: %v", err)
	}

	validator := &ResourceIndexPolicyValidator{
		CelValidator: celValidator,
	}

	policy := &policyv1alpha1.ResourceIndexPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
	}

	warnings, err := validator.ValidateDelete(context.Background(), policy)

	// Delete should always succeed
	if err != nil {
		t.Errorf("ValidateDelete() unexpected error: %v", err)
	}

	if len(warnings) > 0 {
		t.Errorf("ValidateDelete() unexpected warnings: %v", warnings)
	}
}

func TestValidateCreateWithInvalidJSONPath(t *testing.T) {
	celValidator, err := cel.NewValidator(10)
	if err != nil {
		t.Fatalf("failed to create CEL validator: %v", err)
	}

	validator := &ResourceIndexPolicyValidator{
		CelValidator: celValidator,
	}

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid path with bracket notation",
			path:    `.metadata.labels["app"]`,
			wantErr: false,
		},
		{
			name:        "invalid path - missing leading dot",
			path:        "spec.name",
			wantErr:     true,
			errContains: "must start with '.'",
		},
		{
			name:        "invalid path - special chars without brackets",
			path:        ".spec.my-field",
			wantErr:     true,
			errContains: "invalid path syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &policyv1alpha1.ResourceIndexPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: policyv1alpha1.ResourceIndexPolicySpec{
					TargetResource: policyv1alpha1.TargetResource{
						Group:   "contacts.miloapis.com",
						Version: "v1",
						Kind:    "Contact",
					},
					Conditions: []policyv1alpha1.PolicyCondition{
						{
							Name:       "is-active",
							Expression: "status.active == true",
						},
					},
					Fields: []policyv1alpha1.FieldPolicy{
						{
							Path:       tt.path,
							Searchable: true,
						},
					},
				},
			}

			_, err := validator.ValidateCreate(context.Background(), policy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCreate() expected error for path %q but got none", tt.path)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateCreate() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCreate() unexpected error for path %q: %v", tt.path, err)
				}
			}
		})
	}
}
