// Package application_v2 contains the plan logic for KKP applications.
package application_v2

import (
	"github.com/kubermatic/go-kubermatic/models"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Defaults & Validation ----------

// SetDefaults applies default values to the application plan.
func (p *Plan) SetDefaults() {
	if p.Namespace == "" {
		p.Namespace = kkp.DefaultNamespace
	}
}

// Validate validates the application plan configuration.
func (p *Plan) Validate() error {
	if err := kkp.ValidateResourceName(p.Name); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(p.ClusterID, "cluster_id"); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(p.ApplicationName, "application_name"); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(p.ApplicationVersion, "application_version"); err != nil {
		return err
	}
	return nil
}

// ---------- Build ApplicationInstallation spec for KKP API ----------

// ToApplicationInstallation converts the plan to a KKP application installation model.
func (p *Plan) ToApplicationInstallation() (*models.ApplicationInstallation, error) {
	return kkp.ExecuteToModel(p, p.buildApplicationInstallation)
}

func (p *Plan) buildApplicationInstallation() (*models.ApplicationInstallation, error) {
	// Create application reference
	appRef := &models.ApplicationRef{
		Name:    p.ApplicationName,
		Version: p.ApplicationVersion,
	}

	// Create namespace spec
	namespaceSpec := &models.NamespaceSpec{
		Name: p.Namespace,
	}

	// Create spec
	spec := &models.ApplicationInstallationSpec{
		ApplicationRef: appRef,
		Namespace:      namespaceSpec,
	}

	// Add values if provided
	if p.Values != nil {
		spec.Values = models.RawExtension(p.Values)
	}

	return &models.ApplicationInstallation{
		Name:      p.Name,
		Namespace: p.Namespace,
		Spec:      spec,
	}, nil
}
