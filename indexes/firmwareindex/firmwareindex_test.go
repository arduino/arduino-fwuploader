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

package firmwareindex

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestIndexParsing(t *testing.T) {
	// semver.WarnInvalidVersionWhenParsingRelaxed = true
	list, err := paths.New("testdata").ReadDir()
	require.NoError(t, err)
	list.FilterSuffix(".json")
	for _, indexFile := range list {
		t.Logf("testing with index: %s", indexFile)
		index, e := LoadIndexNoSign(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, index)

		index, e = LoadIndex(indexFile)
		require.NoError(t, e)
		require.NotEmpty(t, index)
	}
}
