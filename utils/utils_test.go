package utils

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestGetCompatibleWith(t *testing.T) {
	root, err := paths.Getwd()
	require.NoError(t, err)
	require.NoError(t, root.ToAbs())
	testrunner := func(board string) {
		res := GetCompatibleWith(board, root.String())
		require.NotNil(t, res)
		hasLoader := false
		for _, e := range res {
			for _, i := range e {
				if i.IsLoader {
					require.False(t, hasLoader, "loader must be unique")
					hasLoader = true
					require.NotEmpty(t, i.Name)
					require.NotEmpty(t, i.Path)
				}
			}
		}
		require.True(t, hasLoader, "loader must be present")
	}

	testrunner("mkrwifi1010")
	testrunner("mkr1000")
	testrunner("nano_33_iot")
}
