// Package machine_deployment_v2 contains the plan logic for KKP machine deployments.
package machine_deployment_v2

import (
	"errors"
	"fmt"
	"strings"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	"github.com/kubermatic/go-kubermatic/models"
)

// ---------- Defaults & Validation ----------

// SetDefaults applies default values to the machine deployment plan.
func (p *Plan) SetDefaults() {
	if p.Replicas == 0 {
		p.Replicas = int32(kkp.DefaultReplicas)
	}

	// Set OpenStack defaults
	if p.Cloud == kkp.CloudOpenStack && p.OpenStack != nil {
		if p.OpenStack.DiskSize == 0 {
			p.OpenStack.DiskSize = int32(kkp.DefaultDiskSize)
		}
	}
}

// Validate validates the machine deployment plan configuration.
func (p *Plan) Validate() error {
	if err := kkp.ValidateResourceName(p.Name); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(p.ClusterID, "cluster_id"); err != nil {
		return err
	}
	if err := kkp.ValidateCloudProvider(p.Cloud); err != nil {
		return err
	}
	if err := kkp.ValidateReplicas(int64(p.Replicas)); err != nil {
		return err
	}
	
	// Convert int32 pointers to int64 pointers for validation
	var minReplicas, maxReplicas *int64
	if p.MinReplicas != nil {
		val := int64(*p.MinReplicas)
		minReplicas = &val
	}
	if p.MaxReplicas != nil {
		val := int64(*p.MaxReplicas)
		maxReplicas = &val
	}
	if err := kkp.ValidateAutoscaling(minReplicas, maxReplicas); err != nil {
		return err
	}

	switch p.Cloud {
	case kkp.CloudOpenStack:
		if p.OpenStack == nil {
			return errors.New("openstack block must be set for cloud=openstack")
		}
		return p.validateOpenStack()
	case kkp.CloudAWS:
		if p.AWS == nil {
			return errors.New("aws block must be set for cloud=aws")
		}
		// AWS validation would go here when implemented
	case kkp.CloudVSphere:
		if p.VSphere == nil {
			return errors.New("vsphere block must be set for cloud=vsphere")
		}
		// VSphere validation would go here when implemented
	case kkp.CloudAzure:
		if p.Azure == nil {
			return errors.New("azure block must be set for cloud=azure")
		}
		// Azure validation would go here when implemented
	default:
		return fmt.Errorf("unsupported cloud provider %q", p.Cloud)
	}

	return nil
}

func (p *Plan) validateOpenStack() error {
	os := p.OpenStack
	if err := kkp.ValidateRequiredString(os.Flavor, "openstack.flavor"); err != nil {
		return err
	}
	if err := kkp.ValidateRequiredString(os.Image, "openstack.image"); err != nil {
		return err
	}
	if err := kkp.ValidateDiskSize(int64(os.DiskSize)); err != nil {
		return fmt.Errorf("openstack.%s", err.Error())
	}
	return nil
}

// ---------- Build NodeDeployment spec for KKP API ----------

// ToMachineDeploymentSpec converts the plan to a KKP machine deployment specification.
func (p *Plan) ToMachineDeploymentSpec() (*models.NodeDeployment, error) {
	return kkp.ExecuteToModel(p, p.buildMachineDeploymentSpec)
}

func (p *Plan) buildMachineDeploymentSpec() (*models.NodeDeployment, error) {
	spec := &models.NodeDeploymentSpec{
		Replicas: &p.Replicas,
		Template: &models.NodeSpec{
			Versions: &models.NodeVersionInfo{},
			OperatingSystem: &models.OperatingSystemSpec{
				Ubuntu: &models.UbuntuSpec{
					DistUpgradeOnBoot: false, // Disable automatic dist upgrade
				},
			},
			Labels: map[string]string{
				"system/cluster": p.ClusterID,
				"system/project": "", // Will be filled by KKP
			},
		},
		Paused: false, // Explicitly set to false
	}

	// Set Kubernetes version if specified
	if strings.TrimSpace(p.K8sVersion) != "" {
		spec.Template.Versions.Kubelet = p.K8sVersion
	}

	// Set deployment options from plan
	spec.Paused = p.Paused

	// Configure cloud-specific settings
	switch p.Cloud {
	case kkp.CloudOpenStack:
		os := p.OpenStack
		cloudSpec := &models.OpenstackNodeSpec{
			Flavor:        &os.Flavor,
			Image:         &os.Image,
			UseFloatingIP: os.UseFloatingIP,
		}

		// Set availability zone if specified
		if strings.TrimSpace(os.AvailabilityZone) != "" {
			cloudSpec.AvailabilityZone = os.AvailabilityZone
		}

		// TODO: Disk size configuration - need to check OpenStack node spec fields

		spec.Template.Cloud = &models.NodeCloudSpec{
			Openstack: cloudSpec,
		}

	case "aws":
		// AWS implementation would go here when needed
		spec.Template.Cloud = &models.NodeCloudSpec{
			Aws: &models.AWSNodeSpec{},
		}

	case "vsphere":
		// VSphere implementation would go here when needed
		spec.Template.Cloud = &models.NodeCloudSpec{
			Vsphere: &models.VSphereNodeSpec{},
		}

	case "azure":
		// Azure implementation would go here when needed
		spec.Template.Cloud = &models.NodeCloudSpec{
			Azure: &models.AzureNodeSpec{},
		}
	}

	return &models.NodeDeployment{
		Name: p.Name,
		Spec: spec,
	}, nil
}
