// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package verapdf_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const mplHeaderPrefix = `// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0`

func TestAllVerapdfFilesUseMPLHeader(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "go", "pkg", "verapdf")
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		require.NoError(t, err)
		if d == nil || d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, path := range files {
		data, err := os.ReadFile(path)
		require.NoErrorf(t, err, "read %s", path)
		require.Truef(t, strings.HasPrefix(string(data), mplHeaderPrefix), "missing MPL header in %s", path)
	}
}

func TestVerapdfLicenseFileExistsAndLooksComplete(t *testing.T) {
	licensePath := filepath.Join("..", "..", "..", "..", "go", "pkg", "verapdf", "LICENSE_MPL2")
	data, err := os.ReadFile(licensePath)
	require.NoError(t, err)

	content := string(data)
	require.Contains(t, content, "Mozilla Public License")
	require.Contains(t, content, "Version 2.0")
	require.Greater(t, len(strings.Split(strings.TrimSpace(content), "\n")), 100)
}
