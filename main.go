// Package main implements the Terraform Provider for KKP (Kubermatic Kubernetes Platform).
// This provider enables management of KKP resources including clusters, machine deployments,
// addons, applications, and SSH keys through Terraform.
package main

import (
	"context"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "0.0.1"

func main() {
	err := providerserver.Serve(
		context.Background(),
		provider.New(version),
		providerserver.ServeOpts{
			Address: "registry.opentofu.org/armagankaratosun/kkp",
		},
	)
	if err != nil {
		panic(err)
	}
}
