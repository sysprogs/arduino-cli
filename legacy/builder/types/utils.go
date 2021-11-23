// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package types

// duplication of utils.SliceContains! Thanks golang! Why? Because with golang you can't have import cycles
func sliceContains(slice []string, target string) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}

func sliceContainsSourceFile(slice []SourceFile, target SourceFile) bool {
	for _, elem := range slice {
		if elem.Origin == target.Origin && elem.RelativePath.EqualsTo(target.RelativePath) {
			return true
		}
	}
	return false
}
