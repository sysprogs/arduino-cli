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

package builder_utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func PrintProgressIfProgressEnabledAndMachineLogger(ctx *types.Context) {

	if !ctx.Progress.PrintEnabled {
		return
	}

	log := ctx.GetLogger()
	if log.Name() == "machine" {
		log.Println(constants.LOG_LEVEL_INFO, constants.MSG_PROGRESS, strconv.FormatFloat(float64(ctx.Progress.Progress), 'f', 2, 32))
	}
}

func CompileFilesRecursive(ctx *types.Context, sourcePath *paths.Path, buildPath *paths.Path, buildProperties *properties.Map, includes []string, libraryModel *types.CodeModelLibrary) (paths.PathList, error) {
	objectFiles, err := CompileFiles(ctx, sourcePath, false, buildPath, buildProperties, includes, libraryModel)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	folders, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterDirs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, folder := range folders {
		subFolderObjectFiles, err := CompileFilesRecursive(ctx, sourcePath.Join(folder.Name()), buildPath.Join(folder.Name()), buildProperties, includes, libraryModel)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		objectFiles.AddAll(subFolderObjectFiles)
	}

	return objectFiles, nil
}

func CompileFiles(ctx *types.Context, sourcePath *paths.Path, recurse bool, buildPath *paths.Path, buildProperties *properties.Map, includes []string, libraryModel *types.CodeModelLibrary) (paths.PathList, error) {
	sSources, err := findFilesInFolder(sourcePath, ".S", recurse)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cSources, err := findFilesInFolder(sourcePath, ".c", recurse)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cppSources, err := findFilesInFolder(sourcePath, ".cpp", recurse)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ctx.Progress.AddSubSteps(len(sSources) + len(cSources) + len(cppSources))
	defer ctx.Progress.RemoveSubSteps()

	sObjectFiles, err := compileFilesWithRecipe(ctx, sourcePath, sSources, buildPath, buildProperties, includes, constants.RECIPE_S_PATTERN, libraryModel)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cObjectFiles, err := compileFilesWithRecipe(ctx, sourcePath, cSources, buildPath, buildProperties, includes, constants.RECIPE_C_PATTERN, libraryModel)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cppObjectFiles, err := compileFilesWithRecipe(ctx, sourcePath, cppSources, buildPath, buildProperties, includes, constants.RECIPE_CPP_PATTERN, libraryModel)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	objectFiles := paths.NewPathList()
	objectFiles.AddAll(sObjectFiles)
	objectFiles.AddAll(cObjectFiles)
	objectFiles.AddAll(cppObjectFiles)
	return objectFiles, nil
}

func ReplaceOptimizationFlags(str string) string {
	var tmp = strings.Split(str, " ")
	for k, v := range tmp {
		if v == "-O2" || v == "-Os" || v == "-O1" || v == "-Og" || v == "-O3" {
			tmp[k] = "-O0"
		} else if v == "-flto" {
			tmp[k] = ""
		}
	}

	return strings.Join(tmp, " ")
}

func RemoveOptimizationFromBuildProperties(properties *properties.Map) *properties.Map {
	var result = properties.Clone()

	result.Set("compiler.c.flags", ReplaceOptimizationFlags(result.Get("compiler.c.flags")))
	result.Set("compiler.cpp.flags", ReplaceOptimizationFlags(result.Get("compiler.cpp.flags")))
	result.Set("build.flags.optimize", ReplaceOptimizationFlags(result.Get("build.flags.optimize")))
	return result
}

func ExpandSysprogsExtensionProperties(properties *properties.Map) *properties.Map {
	var result = properties.Clone()

	result.Set("compiler.c.flags", result.Get("compiler.c.flags") + " " + result.Get("com.sysprogs.extraflags"))
	result.Set("compiler.cpp.flags", result.Get("compiler.cpp.flags") + " " + result.Get("com.sysprogs.extraflags"))
	return result
}



func findFilesInFolder(sourcePath *paths.Path, extension string, recurse bool) (paths.PathList, error) {
	files, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterFilesWithExtensions(extension))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var sources paths.PathList
	for _, file := range files {
		sources = append(sources, sourcePath.Join(file.Name()))
	}

	if recurse {
		folders, err := utils.ReadDirFiltered(sourcePath.String(), utils.FilterDirs)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, folder := range folders {
			otherSources, err := findFilesInFolder(sourcePath.Join(folder.Name()), extension, recurse)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			sources = append(sources, otherSources...)
		}
	}

	return sources, nil
}

func findAllFilesInFolder(sourcePath string, recurse bool) ([]string, error) {
	files, err := utils.ReadDirFiltered(sourcePath, utils.FilterFiles())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var sources []string
	for _, file := range files {
		sources = append(sources, filepath.Join(sourcePath, file.Name()))
	}

	if recurse {
		folders, err := utils.ReadDirFiltered(sourcePath, utils.FilterDirs)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, folder := range folders {
			if !utils.IsSCCSOrHiddenFile(folder) {
				// Skip SCCS directories as they do not influence the build and can be very large
				otherSources, err := findAllFilesInFolder(filepath.Join(sourcePath, folder.Name()), recurse)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				sources = append(sources, otherSources...)
			}
		}
	}

	return sources, nil
}

func compileFilesWithRecipe(ctx *types.Context, sourcePath *paths.Path, sources paths.PathList, buildPath *paths.Path, buildProperties *properties.Map, includes []string, recipe string, libraryModel *types.CodeModelLibrary) (paths.PathList, error) {
	objectFiles := paths.NewPathList()
	if len(sources) == 0 {
		return objectFiles, nil
	}
	var objectFilesMux sync.Mutex
	var errorsList []error
	var errorsMux sync.Mutex

	queue := make(chan *paths.Path)
	job := func(source *paths.Path) {
		objectFile, err := compileFileWithRecipe(ctx, sourcePath, source, buildPath, buildProperties, includes, recipe, libraryModel)
		if err != nil {
			errorsMux.Lock()
			errorsList = append(errorsList, err)
			errorsMux.Unlock()
		} else {
			objectFilesMux.Lock()
			objectFiles.Add(objectFile)
			objectFilesMux.Unlock()
		}
	}

	// Spawn jobs runners
	var wg sync.WaitGroup
	jobs := ctx.Jobs
	if jobs == 0 {
		jobs = runtime.NumCPU()
	}
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			for source := range queue {
				job(source)
			}
			wg.Done()
		}()
	}

	// Feed jobs until error or done
	for _, source := range sources {
		errorsMux.Lock()
		gotError := len(errorsList) > 0
		errorsMux.Unlock()
		if gotError {
			break
		}
		queue <- source

		ctx.Progress.CompleteStep()
		PrintProgressIfProgressEnabledAndMachineLogger(ctx)
	}
	close(queue)
	wg.Wait()
	if len(errorsList) > 0 {
		// output the first error
		return nil, errors.WithStack(errorsList[0])
	}
	objectFiles.Sort()
	return objectFiles, nil
}

func compileFileWithRecipe(ctx *types.Context, sourcePath *paths.Path, source *paths.Path, buildPath *paths.Path, buildProperties *properties.Map, includes []string, recipe string, libraryModel *types.CodeModelLibrary) (*paths.Path, error) {
	logger := ctx.GetLogger()
	properties := buildProperties.Clone()
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel))
	properties.Set(constants.BUILD_PROPERTIES_INCLUDES, strings.Join(includes, constants.SPACE))
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, source)
	relativeSource, err := sourcePath.RelTo(source)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	depsFile := buildPath.Join(relativeSource.String() + ".d")
	objectFile := buildPath.Join(relativeSource.String() + ".o")

	properties.SetPath(constants.BUILD_PROPERTIES_OBJECT_FILE, objectFile)
	err = objectFile.Parent().MkdirAll()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	objIsUpToDate, err := ObjFileIsUpToDate(ctx, source, objectFile, depsFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	command, err := PrepareCommandForRecipe(properties, recipe, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if ctx.CompilationDatabase != nil {
		ctx.CompilationDatabase.Add(source, command)
	}

	if libraryModel != nil {
		var invocation = new(types.CodeModelGCCInvocation)
		invocation.GCC = command.Path
		invocation.InputFile = source.String()
		invocation.ObjectFile = properties.Get(constants.BUILD_PROPERTIES_OBJECT_FILE)
		invocation.Arguments = command.Args[1:]
		libraryModel.Invocations = append(libraryModel.Invocations, invocation)
	}

	if !objIsUpToDate && !ctx.OnlyUpdateCompilationDatabase && libraryModel == nil{
		_, _, err = utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else if ctx.Verbose {
		if objIsUpToDate {
			logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_PREVIOUS_COMPILED_FILE, objectFile)
		} else {
			logger.Println("info", "Skipping compile of: {0}", objectFile)
		}
	}

	return objectFile, nil
}

func ObjFileIsUpToDate(ctx *types.Context, sourceFile, objectFile, dependencyFile *paths.Path) (bool, error) {
	logger := ctx.GetLogger()
	debugLevel := ctx.DebugLevel
	if debugLevel >= 20 {
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Checking previous results for {0} (result = {1}, dep = {2})", sourceFile, objectFile, dependencyFile)
	}
	if objectFile == nil || dependencyFile == nil {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: nil")
		}
		return false, nil
	}

	sourceFile = sourceFile.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return false, errors.WithStack(err)
	}

	objectFile = objectFile.Clean()
	objectFileStat, err := objectFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: {0}", objectFile)
			}
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	dependencyFile = dependencyFile.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: {0}", dependencyFile)
			}
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	if sourceFileStat.ModTime().After(objectFileStat.ModTime()) {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "{0} newer than {1}", sourceFile, objectFile)
		}
		return false, nil
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "{0} newer than {1}", sourceFile, dependencyFile)
		}
		return false, nil
	}

	rows, err := dependencyFile.ReadFileAsLines()
	if err != nil {
		return false, errors.WithStack(err)
	}

	rows = utils.Map(rows, removeEndingBackSlash)
	rows = utils.Map(rows, strings.TrimSpace)
	rows = utils.Map(rows, unescapeDep)
	rows = utils.Filter(rows, nonEmptyString)

	if len(rows) == 0 {
		return true, nil
	}

	firstRow := rows[0]
	if !strings.HasSuffix(firstRow, ":") {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "No colon in first line of depfile")
		}
		return false, nil
	}
	objFileInDepFile := firstRow[:len(firstRow)-1]
	if objFileInDepFile != objectFile.String() {
		if debugLevel >= 20 {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Depfile is about different file: {0}", objFileInDepFile)
		}
		return false, nil
	}

	// The first line of the depfile contains the path to the object file to generate.
	// The second line of the depfile contains the path to the source file.
	// All subsequent lines contain the header files necessary to compile the object file.

	// If we don't do this check it might happen that trying to compile a source file
	// that has the same name but a different path wouldn't recreate the object file.
	if sourceFile.String() != strings.Trim(rows[1], " ") {
		return false, nil
	}

	rows = rows[1:]
	for _, row := range rows {
		depStat, err := os.Stat(row)
		if err != nil && !os.IsNotExist(err) {
			// There is probably a parsing error of the dep file
			// Ignore the error and trigger a full rebuild anyway
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Failed to read: {0}", row)
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, err.Error())
			}
			return false, nil
		}
		if os.IsNotExist(err) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "Not found: {0}", row)
			}
			return false, nil
		}
		if depStat.ModTime().After(objectFileStat.ModTime()) {
			if debugLevel >= 20 {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, "{0} newer than {1}", row, objectFile)
			}
			return false, nil
		}
	}

	return true, nil
}

func unescapeDep(s string) string {
	s = strings.Replace(s, "\\ ", " ", -1)
	s = strings.Replace(s, "\\\t", "\t", -1)
	s = strings.Replace(s, "\\#", "#", -1)
	s = strings.Replace(s, "$$", "$", -1)
	s = strings.Replace(s, "\\\\", "\\", -1)
	return s
}

func removeEndingBackSlash(s string) string {
	if strings.HasSuffix(s, "\\") {
		s = s[:len(s)-1]
	}
	return s
}

func nonEmptyString(s string) bool {
	return s != constants.EMPTY_STRING
}

func CoreOrReferencedCoreHasChanged(corePath, targetCorePath, targetFile *paths.Path) bool {

	targetFileStat, err := targetFile.Stat()
	if err == nil {
		files, err := findAllFilesInFolder(corePath.String(), true)
		if err != nil {
			return true
		}
		for _, file := range files {
			fileStat, err := os.Stat(file)
			if err != nil || fileStat.ModTime().After(targetFileStat.ModTime()) {
				return true
			}
		}
		if targetCorePath != nil && !strings.EqualFold(corePath.String(), targetCorePath.String()) {
			return CoreOrReferencedCoreHasChanged(targetCorePath, nil, targetFile)
		}
		return false
	}
	return true
}

func TXTBuildRulesHaveChanged(corePath, targetCorePath, targetFile *paths.Path) bool {

	targetFileStat, err := targetFile.Stat()
	if err == nil {
		files, err := findAllFilesInFolder(corePath.String(), true)
		if err != nil {
			return true
		}
		for _, file := range files {
			// report changes only for .txt files
			if filepath.Ext(file) != ".txt" {
				continue
			}
			fileStat, err := os.Stat(file)
			if err != nil || fileStat.ModTime().After(targetFileStat.ModTime()) {
				return true
			}
		}
		if targetCorePath != nil && !corePath.EqualsTo(targetCorePath) {
			return TXTBuildRulesHaveChanged(targetCorePath, nil, targetFile)
		}
		return false
	}
	return true
}

func ArchiveCompiledFiles(ctx *types.Context, buildPath *paths.Path, archiveFile *paths.Path, objectFilesToArchive paths.PathList, buildProperties *properties.Map, libraryModel *types.CodeModelLibrary) (*paths.Path, error) {
	logger := ctx.GetLogger()
	archiveFilePath := buildPath.JoinPath(archiveFile)

	if ctx.OnlyUpdateCompilationDatabase {
		if ctx.Verbose {
			logger.Println("info", "Skipping archive creation of: {0}", archiveFilePath)
		}
		return archiveFilePath, nil
	}

	if archiveFileStat, err := archiveFilePath.Stat(); err == nil {
		rebuildArchive := false
		for _, objectFile := range objectFilesToArchive {
			objectFileStat, err := objectFile.Stat()
			if err != nil || objectFileStat.ModTime().After(archiveFileStat.ModTime()) {
				// need to rebuild the archive
				rebuildArchive = true
				break
			}
		}

		// something changed, rebuild the core archive
		if rebuildArchive {
			if err := archiveFilePath.Remove(); err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if ctx.Verbose {
				logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_PREVIOUS_COMPILED_FILE, archiveFilePath)
			}
			return archiveFilePath, nil
		}
	}

	for _, objectFile := range objectFilesToArchive {
		properties := buildProperties.Clone()
		properties.Set(constants.BUILD_PROPERTIES_ARCHIVE_FILE, archiveFilePath.Base())
		properties.SetPath(constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH, archiveFilePath)
		properties.SetPath(constants.BUILD_PROPERTIES_OBJECT_FILE, objectFile)

		command, err := PrepareCommandForRecipe(properties, constants.RECIPE_AR_PATTERN, false)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		_, _, err = utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return archiveFilePath, nil
}

const COMMANDLINE_LIMIT = 30000

func PrepareCommandForRecipe(buildProperties *properties.Map, recipe string, removeUnsetProperties bool) (*exec.Cmd, error) {
	pattern := buildProperties.Get(recipe)
	if pattern == "" {
		return nil, errors.Errorf("%s pattern is missing", recipe)
	}

	commandLine := buildProperties.ExpandPropsInString(pattern)
	if removeUnsetProperties {
		commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	}

	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	command := exec.Command(parts[0], parts[1:]...)

	// if the overall commandline is too long for the platform
	// try reducing the length by making the filenames relative
	// and changing working directory to build.path
	if len(commandLine) > COMMANDLINE_LIMIT {
		relativePath := buildProperties.Get("build.path")
		for i, arg := range command.Args {
			if _, err := os.Stat(arg); os.IsNotExist(err) {
				continue
			}
			rel, err := filepath.Rel(relativePath, arg)
			if err == nil && !strings.Contains(rel, "..") && len(rel) < len(arg) {
				command.Args[i] = rel
			}
		}
		command.Dir = relativePath
	}

	return command, nil
}
