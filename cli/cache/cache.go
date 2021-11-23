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

package cache

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCommand created a new `cache` command
func NewCommand() *cobra.Command {
	cacheCommand := &cobra.Command{
		Use:   "cache",
		Short: "Arduino cache commands.",
		Long:  "Arduino cache commands.",
		Example: "# Clean caches.\n" +
			" " + os.Args[0] + " cache clean\n\n",
	}

	cacheCommand.AddCommand(initCleanCommand())

	return cacheCommand
}
