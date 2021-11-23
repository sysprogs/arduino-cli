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

import (
	"io"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
)

type ProgressStruct struct {
	PrintEnabled bool
	Progress     float32
	StepAmount   float32
	Parent       *ProgressStruct
}

func (p *ProgressStruct) AddSubSteps(steps int) {
	p.Parent = &ProgressStruct{
		Progress:   p.Progress,
		StepAmount: p.StepAmount,
		Parent:     p.Parent,
	}
	if p.StepAmount == 0.0 {
		p.StepAmount = 100.0
	}
	p.StepAmount /= float32(steps)
}

func (p *ProgressStruct) RemoveSubSteps() {
	p.Progress = p.Parent.Progress
	p.StepAmount = p.Parent.StepAmount
	p.Parent = p.Parent.Parent
}

func (p *ProgressStruct) CompleteStep() {
	p.Progress += p.StepAmount
}

// Context structure
type Context struct {
	// Build options
	HardwareDirs         paths.PathList
	BuiltInToolsDirs     paths.PathList
	BuiltInLibrariesDirs paths.PathList
	OtherLibrariesDirs   paths.PathList
	LibraryDirs          paths.PathList // List of paths pointing to individual library root folders
	SketchLocation       *paths.Path
	WatchedLocations     paths.PathList
	ArduinoAPIVersion    string
	FQBN                 *cores.FQBN
	CodeCompleteAt       string
	Clean                bool

	// Build options are serialized here
	BuildOptionsJson         string
	BuildOptionsJsonPrevious string

	PackageManager *packagemanager.PackageManager
	Hardware       cores.Packages
	AllTools       []*cores.ToolRelease
	RequiredTools  []*cores.ToolRelease
	TargetBoard    *cores.Board
	TargetPackage  *cores.Package
	TargetPlatform *cores.PlatformRelease
	ActualPlatform *cores.PlatformRelease
	USBVidPid      string

	PlatformKeyRewrites    PlatforKeysRewrite
	HardwareRewriteResults map[*cores.PlatformRelease][]PlatforKeyRewrite

	BuildProperties              *properties.Map
	BuildCore                    string
	BuildPath                    *paths.Path
	BuildCachePath               *paths.Path
	SketchBuildPath              *paths.Path
	CoreBuildPath                *paths.Path
	CoreBuildCachePath           *paths.Path
	CoreArchiveFilePath          *paths.Path
	CoreObjectsFiles             paths.PathList
	LibrariesBuildPath           *paths.Path
	LibrariesObjectFiles         paths.PathList
	PreprocPath                  *paths.Path
	SketchObjectFiles            paths.PathList
	IgnoreSketchFolderNameErrors bool

	CollectedSourceFiles *UniqueSourceFileQueue

	Sketch          *Sketch
	Source          string
	SourceGccMinusE string
	CodeCompletions string

	WarningsLevel string

	// Libraries handling
	LibrariesManager           *librariesmanager.LibrariesManager
	LibrariesResolver          *librariesresolver.Cpp
	ImportedLibraries          libraries.List
	LibrariesResolutionResults map[string]LibraryResolutionResult
	IncludeFolders             paths.PathList
	//OutputGccMinusM            string

	// C++ Parsing
	CTagsOutput                 string
	CTagsTargetFile             *paths.Path
	CTagsOfPreprocessedSource   []*CTag
	LineOffset                  int
	PrototypesSection           string
	PrototypesLineWhereToInsert int
	Prototypes                  []*Prototype
	
	CodeModelBuilder     *CodeModelBuilder
	CodeModelBuilderFile string
	UnoptimizeSketch     bool
	UnoptimizeCore       bool
	UnoptimizeLibraries  bool	

	// Verbosity settings
	Verbose           bool
	DebugPreprocessor bool

	// Compile optimization settings
	OptimizeForDebug  bool
	OptimizationFlags string

	// Dry run, only create progress map
	Progress ProgressStruct

	// Contents of a custom build properties file (line by line)
	CustomBuildProperties []string

	// Logging
	logger     i18n.Logger
	DebugLevel int

	// Reuse old tools since the backing storage didn't change
	CanUseCachedTools bool

	// Experimental: use arduino-preprocessor to create prototypes
	UseArduinoPreprocessor bool

	// Parallel processes
	Jobs int

	// Out and Err stream to redirect all Exec commands
	ExecStdout io.Writer
	ExecStderr io.Writer

	// Sizer results
	ExecutableSectionsSize ExecutablesFileSections

	// Compilation Database to build/update
	CompilationDatabase *builder.CompilationDatabase
	// Set to true to skip build and produce only Compilation Database
	OnlyUpdateCompilationDatabase bool

	// Source code overrides (filename -> content map).
	// The provided source data is used instead of reading it from disk.
	// The keys of the map are paths relative to sketch folder.
	SourceOverride map[string]string
}

// ExecutableSectionSize represents a section of the executable output file
type ExecutableSectionSize struct {
	Name    string
	Size    int
	MaxSize int
}

// ExecutablesFileSections is an array of ExecutablesFileSection
type ExecutablesFileSections []ExecutableSectionSize

// ToRPCExecutableSectionSizeArray transforms this array into a []*rpc.ExecutableSectionSize
func (s ExecutablesFileSections) ToRPCExecutableSectionSizeArray() []*rpc.ExecutableSectionSize {
	res := []*rpc.ExecutableSectionSize{}
	for _, section := range s {
		res = append(res, &rpc.ExecutableSectionSize{
			Name:    section.Name,
			Size:    int64(section.Size),
			MaxSize: int64(section.MaxSize),
		})
	}
	return res
}

func (ctx *Context) ExtractBuildOptions() *properties.Map {
	opts := properties.NewMap()
	opts.Set("hardwareFolders", strings.Join(ctx.HardwareDirs.AsStrings(), ","))
	opts.Set("builtInToolsFolders", strings.Join(ctx.BuiltInToolsDirs.AsStrings(), ","))
	opts.Set("builtInLibrariesFolders", strings.Join(ctx.BuiltInLibrariesDirs.AsStrings(), ","))
	opts.Set("otherLibrariesFolders", strings.Join(ctx.OtherLibrariesDirs.AsStrings(), ","))
	opts.SetPath("sketchLocation", ctx.SketchLocation)
	var additionalFilesRelative []string
	if ctx.Sketch != nil {
		for _, sketch := range ctx.Sketch.AdditionalFiles {
			absPath := ctx.SketchLocation.Parent()
			relPath, err := sketch.Name.RelTo(absPath)
			if err != nil {
				continue // ignore
			}
			additionalFilesRelative = append(additionalFilesRelative, relPath.String())
		}
	}
	opts.Set("fqbn", ctx.FQBN.String())
	opts.Set("runtime.ide.version", ctx.ArduinoAPIVersion)
	opts.Set("customBuildProperties", strings.Join(ctx.CustomBuildProperties, ","))
	opts.Set("additionalFiles", strings.Join(additionalFilesRelative, ","))
	opts.Set("compiler.optimization_flags", ctx.OptimizationFlags)
	return opts
}

func (ctx *Context) InjectBuildOptions(opts *properties.Map) {
	ctx.HardwareDirs = paths.NewPathList(strings.Split(opts.Get("hardwareFolders"), ",")...)
	ctx.BuiltInToolsDirs = paths.NewPathList(strings.Split(opts.Get("builtInToolsFolders"), ",")...)
	ctx.BuiltInLibrariesDirs = paths.NewPathList(strings.Split(opts.Get("builtInLibrariesFolders"), ",")...)
	ctx.OtherLibrariesDirs = paths.NewPathList(strings.Split(opts.Get("otherLibrariesFolders"), ",")...)
	ctx.SketchLocation = opts.GetPath("sketchLocation")
	fqbn, err := cores.ParseFQBN(opts.Get("fqbn"))
	if err != nil {
		i18n.ErrorfWithLogger(ctx.GetLogger(), "Error in FQBN: %s", err)
	}
	ctx.FQBN = fqbn
	ctx.ArduinoAPIVersion = opts.Get("runtime.ide.version")
	ctx.CustomBuildProperties = strings.Split(opts.Get("customBuildProperties"), ",")
	ctx.OptimizationFlags = opts.Get("compiler.optimization_flags")
}

func (ctx *Context) GetLogger() i18n.Logger {
	if ctx.logger == nil {
		return &i18n.HumanLogger{}
	}
	return ctx.logger
}

func (ctx *Context) SetLogger(l i18n.Logger) {
	ctx.logger = l
}
