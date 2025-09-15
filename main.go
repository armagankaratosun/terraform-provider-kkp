// Package main implements the Terraform Provider for KKP (Kubermatic Kubernetes Platform).
// This provider enables management of KKP resources including clusters, machine deployments,
// addons, applications, and SSH keys through Terraform.
package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/provider"
)

var version = "0.0.1"

// providerAddress is the fully qualified provider address used by Terraform/OpenTofu.
// Default points to the Terraform Registry; override at build time for OpenTofu via:
//
//	-ldflags "-X main.providerAddress=registry.opentofu.org/armagankaratosun/kkp"
var providerAddress = "registry.terraform.io/armagankaratosun/kkp"

func main() {
	err := providerserver.Serve(
		context.Background(),
		provider.New(version),
		providerserver.ServeOpts{Address: providerAddress},
	)
	if err != nil {
		panic(err)
	}
}
