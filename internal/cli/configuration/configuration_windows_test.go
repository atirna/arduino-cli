// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You should have been released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software in commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

//go:build windows

package configuration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
)

// https://github.com/arduino/arduino-cli/issues/3276
func TestGetDefaultUserDirInvalidDocumentsShellFolder(t *testing.T) {
	// The "Personal" value of the "User Shell Folders" registry key points
	// to a non-existent directory: the default user folder must fall back to
	// the Documents folder in the user home instead of a relative path.
	fakeHome := t.TempDir()
	missingDocuments := filepath.Join(fakeHome, "Missing", "Documents")
	overridePersonalShellFolder(t, missingDocuments)

	t.Setenv("USERPROFILE", fakeHome)
	require.Equal(t,
		filepath.Join(fakeHome, "Documents", "Arduino"),
		getDefaultUserDir())
}

// overridePersonalShellFolder points the "Personal" value of the "User Shell
// Folders" registry key to the given path for the duration of the test,
// restoring the original value afterwards.
func overridePersonalShellFolder(t *testing.T, documentsPath string) {
	t.Helper()
	const name = "Personal"
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, key.Close()) })
	original, valueType, err := key.GetStringValue(name)
	if err != nil {
		require.ErrorIs(t, err, registry.ErrNotExist)
		t.Cleanup(func() { require.NoError(t, key.DeleteValue(name)) })
	} else {
		t.Cleanup(func() { require.NoError(t, restoreStringValue(key, name, original, valueType)) })
	}
	require.NoError(t, key.SetExpandStringValue(name, documentsPath))
}

func restoreStringValue(key registry.Key, name, value string, valueType uint32) error {
	if valueType == registry.EXPAND_SZ {
		return key.SetExpandStringValue(name, value)
	}
	return key.SetStringValue(name, value)
}
