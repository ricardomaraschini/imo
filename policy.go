package imo

import (
	"github.com/containers/image/v5/signature"
)

// PolicyContext returns the default policy context.
func PolicyContext() (*signature.PolicyContext, error) {
	pol := &signature.Policy{
		Default: signature.PolicyRequirements{
			signature.NewPRInsecureAcceptAnything(),
		},
	}
	return signature.NewPolicyContext(pol)
}
