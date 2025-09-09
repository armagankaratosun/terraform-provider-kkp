package kkp

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateRequiredString validates that a string field is not empty after trimming
func ValidateRequiredString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateCloudProvider validates that the cloud provider is supported
func ValidateCloudProvider(cloud string) error {
	if strings.TrimSpace(cloud) == "" {
		return errors.New("cloud provider is required")
	}

	supportedClouds := []string{"openstack", "aws", "vsphere", "azure"}
	for _, supported := range supportedClouds {
		if cloud == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported cloud provider %q, must be one of: %s",
		cloud, strings.Join(supportedClouds, ", "))
}

// ValidateReplicas validates replica count within acceptable bounds
func ValidateReplicas(replicas int64) error {
	if replicas < 0 {
		return errors.New("replicas cannot be negative")
	}
	if replicas > MaxReplicas {
		return fmt.Errorf("replicas cannot exceed %d", MaxReplicas)
	}
	return nil
}

// ValidateAutoscaling validates autoscaling configuration
func ValidateAutoscaling(minReplicas, maxReplicas *int64) error {
	if minReplicas == nil && maxReplicas == nil {
		return nil // No autoscaling configured
	}

	if minReplicas == nil || maxReplicas == nil {
		return errors.New("both min_replicas and max_replicas must be set when using autoscaling")
	}

	if *minReplicas < 1 {
		return errors.New("min_replicas must be at least 1")
	}
	if *maxReplicas > MaxAutoscalingReplicas {
		return fmt.Errorf("max_replicas cannot exceed %d", MaxAutoscalingReplicas)
	}
	if *minReplicas > *maxReplicas {
		return errors.New("min_replicas cannot be greater than max_replicas")
	}

	return nil
}

// ValidateK8sVersion validates Kubernetes version format
func ValidateK8sVersion(version string) error {
	if strings.TrimSpace(version) == "" {
		return errors.New("k8s_version is required")
	}
	if !K8sVersionPattern.MatchString(version) {
		return fmt.Errorf("k8s_version should look like 1.28 or 1.28.5, got %q", version)
	}
	return nil
}

// ValidateResourceName validates resource name field
func ValidateResourceName(name string) error {
	return ValidateRequiredString(name, "name")
}

// ValidateDiskSize validates disk size within acceptable bounds
func ValidateDiskSize(size int64) error {
	if size < 1 {
		return errors.New("disk_size must be at least 1GB")
	}
	if size > MaxDiskSize {
		return fmt.Errorf("disk_size cannot exceed %dGB", MaxDiskSize)
	}
	return nil
}
