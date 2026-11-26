package properties

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdministratorWriteOnlySpec(t *testing.T) {
	data, err := os.ReadFile("../../specs/device/administrator.yaml")
	require.NoError(t, err)
	spec, err := ParseSpec(data)
	require.NoError(t, err)
	require.True(t, spec.Spec.Params["password-wo"].FinalWriteOnly())
	require.True(t, spec.Spec.Params["password-wo"].FinalSensitive())
	require.Equal(t, "password-wo-version", *spec.Spec.Params["phash"].TerraformProviderConfig.WriteOnlyVersion)
	require.True(t, spec.TerraformProviderConfig.HasCrudHook("PreCreate"))
	require.True(t, spec.TerraformProviderConfig.HasCrudHook("PreUpdate"))
}
