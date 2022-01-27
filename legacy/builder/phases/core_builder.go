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

package phases

import (
	"os"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type CoreBuilder struct{}

func (s *CoreBuilder) Run(ctx *types.Context) error {
	coreBuildPath := ctx.CoreBuildPath
	coreBuildCachePath := ctx.CoreBuildCachePath
	var buildProperties = ctx.BuildProperties

	if err := coreBuildPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}

	if coreBuildCachePath != nil {
		if _, err := coreBuildCachePath.RelTo(ctx.BuildPath); err != nil {
			logger := ctx.GetLogger()
			logger.Println(constants.LOG_LEVEL_INFO, "Couldn't deeply cache core build: {0}", err)
			logger.Println(constants.LOG_LEVEL_INFO, "Running normal build of the core...")
			coreBuildCachePath = nil
			ctx.CoreBuildCachePath = nil
		} else if err := coreBuildCachePath.MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
	}

	var coreModel *types.CodeModelLibrary
	if ctx.CodeModelBuilder != nil {
		coreModel = new(types.CodeModelLibrary)
		ctx.CodeModelBuilder.Core = coreModel
	}

	if ctx.UnoptimizeCore {
		buildProperties = builder_utils.RemoveOptimizationFromBuildProperties(buildProperties)
	}
	
	buildProperties = builder_utils.ExpandSysprogsExtensionProperties(buildProperties, "core")
	
	archiveFile, objectFiles, err := compileCore(ctx, coreBuildPath, coreBuildCachePath, buildProperties, coreModel)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.CoreArchiveFilePath = archiveFile
	ctx.CoreObjectsFiles = objectFiles

	return nil
}

func compileCore(ctx *types.Context, buildPath *paths.Path, buildCachePath *paths.Path, buildProperties *properties.Map, coreModel *types.CodeModelLibrary) (*paths.Path, paths.PathList, error) {
	logger := ctx.GetLogger()
	coreFolder := buildProperties.GetPath("build.core.path")
	variantFolder := buildProperties.GetPath("build.variant.path")

	targetCoreFolder := buildProperties.GetPath(constants.BUILD_PROPERTIES_RUNTIME_PLATFORM_PATH)
	
	if coreModel != nil {
		coreModel.SourceDirectory = coreFolder.String()
		coreModel.Name = buildProperties.Get("name")
	}	

	includes := []string{}
	includes = append(includes, coreFolder.String())
	if variantFolder != nil && variantFolder.IsDir() {
		includes = append(includes, variantFolder.String())
	}
	includes = utils.Map(includes, utils.WrapWithHyphenI)

	var err error

	variantObjectFiles := paths.NewPathList()
	if variantFolder != nil && variantFolder.IsDir() {
		variantObjectFiles, err = builder_utils.CompileFiles(ctx, variantFolder, true, buildPath, buildProperties, includes, coreModel)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	// Recreate the archive if ANY of the core files (including platform.txt) has changed
	realCoreFolder := coreFolder.Parent().Parent()

	var targetArchivedCore *paths.Path
	if buildCachePath != nil {
		archivedCoreName := GetCachedCoreArchiveFileName(buildProperties.Get(constants.BUILD_PROPERTIES_FQBN),
			buildProperties.Get("compiler.optimization_flags"), realCoreFolder)
		targetArchivedCore = buildCachePath.Join(archivedCoreName)
		canUseArchivedCore := !ctx.OnlyUpdateCompilationDatabase &&
			!ctx.Clean &&
			!builder_utils.CoreOrReferencedCoreHasChanged(realCoreFolder, targetCoreFolder, targetArchivedCore)

		if canUseArchivedCore {
			// use archived core
			if ctx.Verbose {
				logger.Println(constants.LOG_LEVEL_INFO, "Using precompiled core: {0}", targetArchivedCore)
			}
			return targetArchivedCore, variantObjectFiles, nil
		}
	}

	coreObjectFiles, err := builder_utils.CompileFiles(ctx, coreFolder, true, buildPath, buildProperties, includes, coreModel)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	archiveFile, err := builder_utils.ArchiveCompiledFiles(ctx, buildPath, paths.New("core.a"), coreObjectFiles, buildProperties, coreModel)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// archive core.a
	if targetArchivedCore != nil && !ctx.OnlyUpdateCompilationDatabase && coreModel != nil {
		err := archiveFile.CopyTo(targetArchivedCore)
		if ctx.Verbose {
			if err == nil {
				logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_ARCHIVING_CORE_CACHE, targetArchivedCore)
			} else if os.IsNotExist(err) {
				logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_CORE_CACHE_UNAVAILABLE, ctx.ActualPlatform)
			} else {
				logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_ERROR_ARCHIVING_CORE_CACHE, targetArchivedCore, err)
			}
		}
	}

	return archiveFile, variantObjectFiles, nil
}

// GetCachedCoreArchiveFileName returns the filename to be used to store
// the global cached core.a.
func GetCachedCoreArchiveFileName(fqbn string, optimizationFlags string, coreFolder *paths.Path) string {
	fqbnToUnderscore := strings.Replace(fqbn, ":", "_", -1)
	fqbnToUnderscore = strings.Replace(fqbnToUnderscore, "=", "_", -1)
	if absCoreFolder, err := coreFolder.Abs(); err == nil {
		coreFolder = absCoreFolder
	} // silently continue if absolute path can't be detected
	hash := utils.MD5Sum([]byte(coreFolder.String() + optimizationFlags))
	realName := "core_" + fqbnToUnderscore + "_" + hash + ".a"
	if len(realName) > 100 {
		// avoid really long names, simply hash the final part
		realName = "core_" + utils.MD5Sum([]byte(fqbnToUnderscore+"_"+hash)) + ".a"
	}
	return realName
}
