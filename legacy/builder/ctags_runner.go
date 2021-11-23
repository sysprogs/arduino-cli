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

package builder

import (
	"os/exec"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/ctags"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type CTagsRunner struct{}

func (s *CTagsRunner) Run(ctx *types.Context) error {
	buildProperties := ctx.BuildProperties
	ctagsTargetFilePath := ctx.CTagsTargetFile

	ctagsProperties := buildProperties.Clone()
	ctagsProperties.Merge(buildProperties.SubTree(constants.BUILD_PROPERTIES_TOOLS_KEY).SubTree(constants.CTAGS))
	ctagsProperties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, ctagsTargetFilePath)

	pattern := ctagsProperties.Get(constants.BUILD_PROPERTIES_PATTERN)
	if pattern == constants.EMPTY_STRING {
		return errors.Errorf("%s pattern is missing", constants.CTAGS)
	}

	commandLine := ctagsProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return errors.WithStack(err)
	}
	command := exec.Command(parts[0], parts[1:]...)
	sourceBytes, _, err := utils.ExecCommand(ctx, command, utils.Capture /* stdout */, utils.Ignore /* stderr */)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.CTagsOutput = string(sourceBytes)

	parser := &ctags.CTagsParser{}

	ctx.CTagsOfPreprocessedSource = parser.Parse(ctx.CTagsOutput, ctx.Sketch.MainFile.Name)
	parser.FixCLinkageTagsDeclarations(ctx.CTagsOfPreprocessedSource)

	protos, line := parser.GeneratePrototypes()
	if line != -1 {
		ctx.PrototypesLineWhereToInsert = line
	}
	ctx.Prototypes = protos

	return nil
}
