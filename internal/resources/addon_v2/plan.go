// Package addon_v2 contains the plan logic for KKP addons.
package addon_v2

import (
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	"github.com/kubermatic/go-kubermatic/models"
)

// ---------- Defaults & Validation ----------

// SetDefaults applies default values to the addon plan.
func (p *Plan) SetDefaults() {
	// No specific defaults for addons currently
}

// Validate validates the addon plan configuration.
func (p *Plan) Validate() error {
	if err := kkp.ValidateResourceName(p.Name); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(p.ClusterID, "cluster_id"); err != nil {
		return err
	}
	return nil
}

// ---------- Build Addon spec for KKP API ----------

// ToAddon converts the plan to a KKP addon model.
func (p *Plan) ToAddon() (*models.Addon, error) {
	return kkp.ExecuteToModel(p, p.buildAddon)
}

func (p *Plan) buildAddon() (*models.Addon, error) {
	spec := &models.AddonSpec{
		ContinuouslyReconcile: p.ContinuouslyReconcile,
		IsDefault:             p.IsDefault,
	}

	// Only set Variables if it's not nil and can be converted to map[string]interface{}
	if p.Variables != nil {
		if vars, ok := p.Variables.(map[string]interface{}); ok {
			spec.Variables = vars
		}
	}

	return &models.Addon{
		Name: p.Name,
		Spec: spec,
	}, nil
}
