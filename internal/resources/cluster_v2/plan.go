// Package cluster_v2 contains the plan logic for KKP clusters.
package cluster_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kubermatic/go-kubermatic/models"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Defaults & Validation ----------

// SetDefaults applies default values to the cluster plan.
func (p *Plan) SetDefaults() {
	if strings.TrimSpace(p.K8sVersion) == "" {
		p.K8sVersion = kkp.DefaultK8sVersion
	}
	if strings.TrimSpace(p.CNI.Type) == "" {
		p.CNI.Type = kkp.DefaultCNIType
	}
	if strings.TrimSpace(p.CNI.Version) == "" {
		p.CNI.Version = kkp.DefaultCNIVersion
	}
	// Only default UseToken=true when using a preset path.
	if p.Cloud == kkp.CloudOpenStack && p.OpenStack != nil && p.Preset != "" {
		if !p.OpenStack.UseToken {
			p.OpenStack.UseToken = true
		}
	}
}

// Validate validates the cluster plan configuration.
func (p *Plan) Validate() error {
	if err := kkp.ValidateResourceName(p.Name); err != nil {
		return err
	}
	if err := kkp.ValidateK8sVersion(p.K8sVersion); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(p.Datacenter, "datacenter"); err != nil {
		return err
	}
	if err := kkp.ValidateCloudProvider(p.Cloud); err != nil {
		return err
	}

	return p.validateCloudConfig()
}

func (p *Plan) validateCloudConfig() error {
	usingPreset := strings.TrimSpace(p.Preset) != ""

	// Generic validation: cloud blocks are optional with preset, required without
	switch p.Cloud {
	case kkp.CloudOpenStack:
		if p.OpenStack == nil && !usingPreset {
			return fmt.Errorf("%s block must be set when not using preset", p.Cloud)
		}
		if p.OpenStack != nil {
			return p.validateOpenStackConfig()
		}
	case kkp.CloudAWS:
		if p.AWS == nil && !usingPreset {
			return fmt.Errorf("%s block must be set when not using preset", p.Cloud)
		}
	case kkp.CloudVSphere:
		if p.VSphere == nil && !usingPreset {
			return fmt.Errorf("%s block must be set when not using preset", p.Cloud)
		}
	case kkp.CloudAzure:
		if p.Azure == nil && !usingPreset {
			return fmt.Errorf("%s block must be set when not using preset", p.Cloud)
		}
	}
	return nil
}

func (p *Plan) validateOpenStackConfig() error {
	appID := strings.TrimSpace(p.OpenStack.ApplicationCredentialID)
	appSecret := strings.TrimSpace(p.OpenStack.ApplicationCredentialSecret)
	usingPreset := strings.TrimSpace(p.Preset) != ""

	// Disallow mixing preset + app creds
	if usingPreset && (appID != "" || appSecret != "") {
		return fmt.Errorf("either set preset OR application_credential_id/_secret, not both")
	}

	// If no preset, require app creds + minimal networking
	if !usingPreset {
		if appID == "" || appSecret == "" {
			return fmt.Errorf("application_credential_id and application_credential_secret are required when no preset is set")
		}
		if err := kkp.ValidateRequiredString(p.OpenStack.Network, "openstack.network"); err != nil {
			return fmt.Errorf("openstack.network is required when no preset is set")
		}
		if err := kkp.ValidateRequiredString(p.OpenStack.SubnetID, "openstack.subnet_id"); err != nil {
			return fmt.Errorf("openstack.subnet_id is required when no preset is set")
		}
		if err := kkp.ValidateRequiredString(p.OpenStack.FloatingIPPool, "openstack.floating_ip_pool"); err != nil {
			return fmt.Errorf("openstack.floating_ip_pool is required when no preset is set")
		}
		if err := kkp.ValidateRequiredString(p.OpenStack.SecurityGroups, "openstack.security_groups"); err != nil {
			return fmt.Errorf("openstack.security_groups must contain at least one security group when no preset is set")
		}
	}

	return nil
}

// ---------- Build CreateClusterSpec for V2 ----------

// ToCreateSpec converts the plan to a KKP cluster create specification.
func (p *Plan) ToCreateSpec(ctx context.Context) (*models.CreateClusterSpec, error) {
	return kkp.ExecuteToModel(p, func() (*models.CreateClusterSpec, error) {
		return p.buildCreateSpec(ctx)
	})
}

func (p *Plan) buildCreateSpec(ctx context.Context) (*models.CreateClusterSpec, error) {
	type looseSpec struct {
		Cluster struct {
			Name       string `json:"name"`
			Credential string `json:"credential,omitempty"`
			Spec       struct {
				Version   string `json:"version"`
				CNIPlugin struct {
					Type    string `json:"type"`
					Version string `json:"version"`
				} `json:"cniPlugin"`
				Cloud struct {
					//  include all common spellings seen across KKP versions/builds, i know this is a mess.
					DatacenterName string `json:"datacenterName,omitempty"`
					Datacenter     string `json:"datacenter,omitempty"`
					DC             string `json:"dc,omitempty"`

					Openstack *struct {
						UseToken                    bool   `json:"useToken,omitempty"`
						ApplicationCredentialID     string `json:"applicationCredentialID,omitempty"`
						ApplicationCredentialSecret string `json:"applicationCredentialSecret,omitempty"`
						Domain                      string `json:"domain,omitempty"`
						Network                     string `json:"network,omitempty"`
						SecurityGroups              string `json:"securityGroups,omitempty"`
						SubnetID                    string `json:"subnetID,omitempty"`
						FloatingIPPool              string `json:"floatingIPPool,omitempty"`
					} `json:"openstack,omitempty"`

					Aws     *struct{} `json:"aws,omitempty"`
					Vsphere *struct{} `json:"vsphere,omitempty"`
					Azure   *struct{} `json:"azure,omitempty"`
				} `json:"cloud"`
			} `json:"spec"`
		} `json:"cluster"`
	}

	var ls looseSpec
	ls.Cluster.Credential = strings.TrimSpace(p.Preset)
	ls.Cluster.Name = p.Name
	ls.Cluster.Spec.Version = p.K8sVersion
	ls.Cluster.Spec.CNIPlugin.Type = p.CNI.Type
	ls.Cluster.Spec.CNIPlugin.Version = p.CNI.Version
	// another mess, related with the first mess above
	ls.Cluster.Spec.Cloud.DatacenterName = p.Datacenter
	ls.Cluster.Spec.Cloud.Datacenter = p.Datacenter
	ls.Cluster.Spec.Cloud.DC = p.Datacenter
	// Set cloud provider configuration based on preset vs app credentials
	switch p.Cloud {
	case kkp.CloudOpenStack:
		if ls.Cluster.Credential == "" {
			// App credentials path: full OpenStack config
			los := &struct {
				UseToken                    bool   `json:"useToken,omitempty"`
				ApplicationCredentialID     string `json:"applicationCredentialID,omitempty"`
				ApplicationCredentialSecret string `json:"applicationCredentialSecret,omitempty"`
				Domain                      string `json:"domain,omitempty"`
				Network                     string `json:"network,omitempty"`
				SecurityGroups              string `json:"securityGroups,omitempty"`
				SubnetID                    string `json:"subnetID,omitempty"`
				FloatingIPPool              string `json:"floatingIPPool,omitempty"`
			}{}

			os := p.OpenStack
			los.ApplicationCredentialID = strings.TrimSpace(os.ApplicationCredentialID)
			los.ApplicationCredentialSecret = strings.TrimSpace(os.ApplicationCredentialSecret)

			if s := strings.TrimSpace(os.Domain); s != "" {
				los.Domain = s
			}
			if s := strings.TrimSpace(os.Network); s != "" {
				los.Network = s
			}
			if s := strings.TrimSpace(os.SubnetID); s != "" {
				los.SubnetID = s
			}
			if s := strings.TrimSpace(os.FloatingIPPool); s != "" {
				los.FloatingIPPool = s
			}
			if s := strings.TrimSpace(os.SecurityGroups); s != "" {
				los.SecurityGroups = s
			}
			ls.Cluster.Spec.Cloud.Openstack = los
		} else {
			// Preset path: empty openstack block to indicate provider type
			ls.Cluster.Spec.Cloud.Openstack = &struct {
				UseToken                    bool   `json:"useToken,omitempty"`
				ApplicationCredentialID     string `json:"applicationCredentialID,omitempty"`
				ApplicationCredentialSecret string `json:"applicationCredentialSecret,omitempty"`
				Domain                      string `json:"domain,omitempty"`
				Network                     string `json:"network,omitempty"`
				SecurityGroups              string `json:"securityGroups,omitempty"`
				SubnetID                    string `json:"subnetID,omitempty"`
				FloatingIPPool              string `json:"floatingIPPool,omitempty"`
			}{}
		}

	case kkp.CloudAWS:
		ls.Cluster.Spec.Cloud.Aws = &struct{}{}
	case kkp.CloudVSphere:
		ls.Cluster.Spec.Cloud.Vsphere = &struct{}{}
	case kkp.CloudAzure:
		ls.Cluster.Spec.Cloud.Azure = &struct{}{}
	}

	raw, _ := json.Marshal(&ls)

	tflog.Debug(ctx, "KKP API request payload", map[string]any{
		"json":   string(raw),
		"preset": p.Preset,
		"cloud":  p.Cloud,
	})

	var spec models.CreateClusterSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("internal JSON bridge: %w", err)
	}
	return &spec, nil
}
