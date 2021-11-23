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

package test

import (
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

// TODO
// I can't find a command I can run on linux, mac and windows
// and that allows to test if the recipe is actually run
// So this test is pretty useless
func TestRecipeRunner(t *testing.T) {
	ctx := &types.Context{}
	buildProperties := properties.NewMap()
	ctx.BuildProperties = buildProperties

	buildProperties.Set("recipe.hooks.prebuild.1.pattern", "echo")

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.RecipeByPrefixSuffixRunner{Prefix: constants.HOOKS_PREBUILD, Suffix: constants.HOOKS_PATTERN_SUFFIX},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
}

func TestRecipesComposition(t *testing.T) {
	require.Equal(t, "recipe.hooks.core.postbuild", constants.HOOKS_CORE_POSTBUILD)
	require.Equal(t, "recipe.hooks.postbuild", constants.HOOKS_POSTBUILD)
	require.Equal(t, "recipe.hooks.linking.prelink", constants.HOOKS_LINKING_PRELINK)
	require.Equal(t, "recipe.hooks.objcopy.preobjcopy", constants.HOOKS_OBJCOPY_PREOBJCOPY)
}
