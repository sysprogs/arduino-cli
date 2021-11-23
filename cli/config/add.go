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

package config

import (
	"os"
	"reflect"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/spf13/cobra"
)

func initAddCommand() *cobra.Command {
	addCommand := &cobra.Command{
		Use:   "add",
		Short: "Adds one or more values to a setting.",
		Long:  "Adds one or more values to a setting.",
		Example: "" +
			"  " + os.Args[0] + " config add board_manager.additional_urls https://example.com/package_example_index.json\n" +
			"  " + os.Args[0] + " config add board_manager.additional_urls https://example.com/package_example_index.json https://another-url.com/package_another_index.json\n",
		Args: cobra.MinimumNArgs(2),
		Run:  runAddCommand,
	}
	return addCommand
}

func runAddCommand(cmd *cobra.Command, args []string) {
	key := args[0]
	kind, err := typeOf(key)
	if err != nil {
		feedback.Error(err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if kind != reflect.Slice {
		feedback.Errorf("The key '%v' is not a list of items, can't add to it.\nMaybe use 'config set'?", key)
		os.Exit(errorcodes.ErrGeneric)
	}

	v := configuration.Settings.GetStringSlice(key)
	v = append(v, args[1:]...)
	configuration.Settings.Set(key, v)

	if err := configuration.Settings.WriteConfig(); err != nil {
		feedback.Errorf("Can't write config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
