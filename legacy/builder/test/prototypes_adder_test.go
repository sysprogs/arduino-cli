// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
// Copyright 2015 Matthijs Kooijman
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
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestPrototypesAdderBridgeExample(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	ctx.DebugLevel = 10

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 33 "+quotedSketchLocation+"\nvoid setup();\n#line 46 "+quotedSketchLocation+"\nvoid loop();\n#line 62 "+quotedSketchLocation+"\nvoid process(BridgeClient client);\n#line 82 "+quotedSketchLocation+"\nvoid digitalCommand(BridgeClient client);\n#line 109 "+quotedSketchLocation+"\nvoid analogCommand(BridgeClient client);\n#line 149 "+quotedSketchLocation+"\nvoid modeCommand(BridgeClient client);\n#line 33 "+quotedSketchLocation+"\n", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchWithIfDef(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("SketchWithIfDef", "SketchWithIfDef.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("SketchWithIfDef", "SketchWithIfDef.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderBaladuino(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("Baladuino", "Baladuino.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("Baladuino", "Baladuino.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderCharWithEscapedDoubleQuote(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("CharWithEscapedDoubleQuote", "CharWithEscapedDoubleQuote.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("CharWithEscapedDoubleQuote", "CharWithEscapedDoubleQuote.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderIncludeBetweenMultilineComment(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("IncludeBetweenMultilineComment", "IncludeBetweenMultilineComment.ino"),
		FQBN:                 parseFQBN(t, "arduino:sam:arduino_due_x_dbg"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("IncludeBetweenMultilineComment", "IncludeBetweenMultilineComment.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderLineContinuations(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("LineContinuations", "LineContinuations.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("LineContinuations", "LineContinuations.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderStringWithComment(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("StringWithComment", "StringWithComment.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("StringWithComment", "StringWithComment.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderSketchWithStruct(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("SketchWithStruct", "SketchWithStruct.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	preprocessed := LoadAndInterpolate(t, filepath.Join("SketchWithStruct", "SketchWithStruct.preprocessed.txt"), ctx)
	obtained := strings.Replace(ctx.Source, "\r\n", "\n", -1)
	// ctags based preprocessing removes the space after "dostuff", but this is still OK
	// TODO: remove this exception when moving to a more powerful parser
	preprocessed = strings.Replace(preprocessed, "void dostuff (A_NEW_TYPE * bar);", "void dostuff(A_NEW_TYPE * bar);", 1)
	obtained = strings.Replace(obtained, "void dostuff (A_NEW_TYPE * bar);", "void dostuff(A_NEW_TYPE * bar);", 1)
	require.Equal(t, preprocessed, obtained)
}

func TestPrototypesAdderSketchWithConfig(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_config", "sketch_with_config.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 13 "+quotedSketchLocation+"\nvoid setup();\n#line 17 "+quotedSketchLocation+"\nvoid loop();\n#line 13 "+quotedSketchLocation+"\n", ctx.PrototypesSection)

	preprocessed := LoadAndInterpolate(t, filepath.Join("sketch_with_config", "sketch_with_config.preprocessed.txt"), ctx)
	require.Equal(t, preprocessed, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderSketchNoFunctionsTwoFiles(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_no_functions_two_files", "sketch_no_functions_two_files.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("sketch_no_functions_two_files", "sketch_no_functions_two_files.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchNoFunctions(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("sketch_no_functions", "sketch_no_functions.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	sketchLocation := paths.New("sketch_no_functions", "sketch_no_functions.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchWithDefaultArgs(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_default_args", "sketch_with_default_args.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 4 "+quotedSketchLocation+"\nvoid setup();\n#line 7 "+quotedSketchLocation+"\nvoid loop();\n#line 1 "+quotedSketchLocation+"\n", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchWithInlineFunction(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_inline_function", "sketch_with_inline_function.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")

	expected := "#line 1 " + quotedSketchLocation + "\nvoid setup();\n#line 2 " + quotedSketchLocation + "\nvoid loop();\n#line 4 " + quotedSketchLocation + "\nshort unsigned int testInt();\n#line 8 " + quotedSketchLocation + "\nstatic int8_t testInline();\n#line 12 " + quotedSketchLocation + "\n__attribute__((always_inline)) uint8_t testAttribute();\n#line 1 " + quotedSketchLocation + "\n"
	obtained := ctx.PrototypesSection
	// ctags based preprocessing removes "inline" but this is still OK
	// TODO: remove this exception when moving to a more powerful parser
	expected = strings.Replace(expected, "static inline int8_t testInline();", "static int8_t testInline();", -1)
	obtained = strings.Replace(obtained, "static inline int8_t testInline();", "static int8_t testInline();", -1)
	// ctags based preprocessing removes "__attribute__ ....." but this is still OK
	// TODO: remove this exception when moving to a more powerful parser
	expected = strings.Replace(expected, "__attribute__((always_inline)) uint8_t testAttribute();", "uint8_t testAttribute();", -1)
	obtained = strings.Replace(obtained, "__attribute__((always_inline)) uint8_t testAttribute();", "uint8_t testAttribute();", -1)
	require.Equal(t, expected, obtained)
}

func TestPrototypesAdderSketchWithFunctionSignatureInsideIFDEF(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_function_signature_inside_ifdef", "sketch_with_function_signature_inside_ifdef.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 1 "+quotedSketchLocation+"\nvoid setup();\n#line 3 "+quotedSketchLocation+"\nvoid loop();\n#line 15 "+quotedSketchLocation+"\nint8_t adalight();\n#line 1 "+quotedSketchLocation+"\n", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchWithUSBCON(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_usbcon", "sketch_with_usbcon.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 5 "+quotedSketchLocation+"\nvoid ciao();\n#line 10 "+quotedSketchLocation+"\nvoid setup();\n#line 15 "+quotedSketchLocation+"\nvoid loop();\n#line 5 "+quotedSketchLocation+"\n", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchWithTypename(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_typename", "sketch_with_typename.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInLibrariesDirs: paths.NewPathList("libraries", "downloaded_libraries"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:leonardo"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	expected := "#line 6 " + quotedSketchLocation + "\nvoid setup();\n#line 10 " + quotedSketchLocation + "\nvoid loop();\n#line 12 " + quotedSketchLocation + "\ntypename Foo<char>::Bar func();\n#line 6 " + quotedSketchLocation + "\n"
	obtained := ctx.PrototypesSection
	// ctags based preprocessing ignores line with typename
	// TODO: remove this exception when moving to a more powerful parser
	expected = strings.Replace(expected, "#line 12 "+quotedSketchLocation+"\ntypename Foo<char>::Bar func();\n", "", -1)
	obtained = strings.Replace(obtained, "#line 12 "+quotedSketchLocation+"\ntypename Foo<char>::Bar func();\n", "", -1)
	require.Equal(t, expected, obtained)
}

func TestPrototypesAdderSketchWithIfDef2(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_ifdef", "sketch_with_ifdef.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:yun"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 5 "+quotedSketchLocation+"\nvoid elseBranch();\n#line 9 "+quotedSketchLocation+"\nvoid f1();\n#line 10 "+quotedSketchLocation+"\nvoid f2();\n#line 12 "+quotedSketchLocation+"\nvoid setup();\n#line 14 "+quotedSketchLocation+"\nvoid loop();\n#line 5 "+quotedSketchLocation+"\n", ctx.PrototypesSection)

	expectedSource := LoadAndInterpolate(t, filepath.Join("sketch_with_ifdef", "sketch.preprocessed.txt"), ctx)
	require.Equal(t, expectedSource, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderSketchWithIfDef2SAM(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_ifdef", "sketch_with_ifdef.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:sam:arduino_due_x_dbg"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 2 "+quotedSketchLocation+"\nvoid ifBranch();\n#line 9 "+quotedSketchLocation+"\nvoid f1();\n#line 10 "+quotedSketchLocation+"\nvoid f2();\n#line 12 "+quotedSketchLocation+"\nvoid setup();\n#line 14 "+quotedSketchLocation+"\nvoid loop();\n#line 2 "+quotedSketchLocation+"\n", ctx.PrototypesSection)

	expectedSource := LoadAndInterpolate(t, filepath.Join("sketch_with_ifdef", "sketch.preprocessed.SAM.txt"), ctx)
	require.Equal(t, expectedSource, strings.Replace(ctx.Source, "\r\n", "\n", -1))
}

func TestPrototypesAdderSketchWithConst(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	sketchLocation := paths.New("sketch_with_const", "sketch_with_const.ino")
	quotedSketchLocation := utils.QuoteCppPath(Abs(t, sketchLocation))

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "#include <Arduino.h>\n#line 1 "+quotedSketchLocation+"\n")
	require.Equal(t, "#line 1 "+quotedSketchLocation+"\nvoid setup();\n#line 2 "+quotedSketchLocation+"\nvoid loop();\n#line 4 "+quotedSketchLocation+"\nconst __FlashStringHelper* test();\n#line 6 "+quotedSketchLocation+"\nconst int test3();\n#line 8 "+quotedSketchLocation+"\nvolatile __FlashStringHelper* test2();\n#line 10 "+quotedSketchLocation+"\nvolatile int test4();\n#line 1 "+quotedSketchLocation+"\n", ctx.PrototypesSection)
}

func TestPrototypesAdderSketchWithDosEol(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       paths.New("eol_processing", "eol_processing.ino"),
		FQBN:                 parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
	// only requires no error as result
}

func TestPrototypesAdderSketchWithSubstringFunctionMember(t *testing.T) {
	DownloadCoresAndToolsAndLibraries(t)
	sketchLocation := paths.New("sketch_with_class_and_method_substring", "sketch_with_class_and_method_substring.ino")
	quotedSketchLocation := utils.QuoteCppString(Abs(t, sketchLocation).String())

	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.NewPathList("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
		SketchLocation:       sketchLocation,
		FQBN:                 parseFQBN(t, "arduino:avr:uno"),
		ArduinoAPIVersion:    "10600",
		Verbose:              true,
	}

	buildPath := SetupBuildPath(t, ctx)
	defer buildPath.RemoveAll()

	commands := []types.Command{

		&builder.ContainerSetupHardwareToolsLibsSketchAndProps{},

		&builder.ContainerMergeCopySketchFiles{},

		&builder.ContainerFindIncludes{},

		&builder.PrintUsedLibrariesIfVerbose{},
		&builder.WarnAboutArchIncompatibleLibraries{},

		&builder.ContainerAddPrototypes{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Contains(t, ctx.Source, "class Foo {\nint blooper(int x) { return x+1; }\n};\n\nFoo foo;\n\n#line 7 "+quotedSketchLocation+"\nvoid setup();")
}
