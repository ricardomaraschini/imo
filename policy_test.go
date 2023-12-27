package imo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_policyContext(t *testing.T) {
	policyCtx, err := policyContext()
	require.NoError(t, err, "PolicyContext should not return an error")
	require.NotNil(t, policyCtx, "PolicyContext should return a non-nil PolicyContext")
}
