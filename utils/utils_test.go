/*
  FirmwareUploader
  Copyright (c) 2021 Arduino LLC.  All right reserved.

  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.

  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
  Lesser General Public License for more details.

  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
*/

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
		t.Run(board, func(t *testing.T) {
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
		})
	}

	testrunner("mkrwifi1010")
	testrunner("mkr1000")
	testrunner("nano_33_iot")
}
