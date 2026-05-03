package imo

import (
	"go.podman.io/image/v5/signature"
)

// policyContext returns the default policy context.
func policyContext() (*signature.PolicyContext, error) {
	pol := &signature.Policy{
		Default: signature.PolicyRequirements{
			signature.NewPRInsecureAcceptAnything(),
		},
	}
	return signature.NewPolicyContext(pol)
}
