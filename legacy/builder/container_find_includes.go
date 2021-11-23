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

/*

Include detection

This code is responsible for figuring out what libraries the current
sketch needs an populates both Context.ImportedLibraries with a list of
Library objects, and Context.IncludeFolders with a list of folders to
put on the include path.

Simply put, every #include in a source file pulls in the library that
provides that source file. This includes source files in the selected
libraries, so libraries can recursively include other libraries as well.

To implement this, the gcc preprocessor is used. A queue is created
containing, at first, the source files in the sketch. Each of the files
in the queue is processed in turn by running the preprocessor on it. If
the preprocessor provides an error, the output is examined to see if the
error is a missing header file originating from a #include directive.

The filename is extracted from that #include directive, and a library is
found that provides it. If multiple libraries provide the same file, one
is slected (how this selection works is not described here, see the
ResolveLibrary function for that). The library selected in this way is
added to the include path through Context.IncludeFolders and the list of
libraries to include in the link through Context.ImportedLibraries.

Furthermore, all of the library source files are added to the queue, to
be processed as well. When the preprocessor completes without showing an
#include error, processing of the file is complete and it advances to
the next. When no library can be found for a included filename, an error
is shown and the process is aborted.

Caching

Since this process is fairly slow (requiring at least one invocation of
the preprocessor per source file), its results are cached.

Just caching the complete result (i.e. the resulting list of imported
libraries) seems obvious, but such a cache is hard to invalidate. Making
a list of all the source and header files used to create the list and
check if any of them changed is probably feasible, but this would also
require caching the full list of libraries to invalidate the cache when
the include to library resolution might have a different result. Another
downside of a complete cache is that any changes requires re-running
everything, even if no includes were actually changed.

Instead, caching happens by keeping a sort of "journal" of the steps in
the include detection, essentially tracing each file processed and each
include path entry added. The cache is used by retracing these steps:
The include detection process is executed normally, except that instead
of running the preprocessor, the include filenames are (when possible)
read from the cache. Then, the include file to library resolution is
again executed normally. The results are checked against the cache and
as long as the results match, the cache is considered valid.

When a source file (or any of the files it includes, as indicated by the
.d file) is changed, the preprocessor is executed as normal for the
file, ignoring any includes from the cache. This does not, however,
invalidate the cache: If the results from the preprocessor match the
entries in the cache, the cache remains valid and can again be used for
the next (unchanged) file.

The cache file uses the JSON format and contains a list of entries. Each
entry represents a discovered library and contains:
 - Sourcefile: The source file that the include was found in
 - Include: The included filename found
 - Includepath: The addition to the include path

There are also some special entries:
 - When adding the initial include path entries, such as for the core
   and variant paths.  These are not discovered, so the Sourcefile and
   Include fields will be empty.
 - When a file contains no (more) missing includes, an entry with an
   empty Include and IncludePath is generated.

*/

package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

type ContainerFindIncludes struct{}

func (s *ContainerFindIncludes) Run(ctx *types.Context) error {
	cachePath := ctx.BuildPath.Join("includes.cache")
	cache := readCache(cachePath)

	appendIncludeFolder(ctx, cache, nil, "", ctx.BuildProperties.GetPath("build.core.path"))
	if ctx.BuildProperties.Get("build.variant.path") != "" {
		appendIncludeFolder(ctx, cache, nil, "", ctx.BuildProperties.GetPath("build.variant.path"))
	}

	sketch := ctx.Sketch
	mergedfile, err := types.MakeSourceFile(ctx, sketch, paths.New(sketch.MainFile.Name.Base()+".cpp"))
	if err != nil {
		return errors.WithStack(err)
	}
	ctx.CollectedSourceFiles.Push(mergedfile)

	sourceFilePaths := ctx.CollectedSourceFiles
	queueSourceFilesFromFolder(ctx, sourceFilePaths, sketch, ctx.SketchBuildPath, false /* recurse */)
	srcSubfolderPath := ctx.SketchBuildPath.Join("src")
	if srcSubfolderPath.IsDir() {
		queueSourceFilesFromFolder(ctx, sourceFilePaths, sketch, srcSubfolderPath, true /* recurse */)
	}

	for !sourceFilePaths.Empty() {
		err := findIncludesUntilDone(ctx, cache, sourceFilePaths.Pop())
		if err != nil {
			cachePath.Remove()
			return errors.WithStack(err)
		}
	}

	// Finalize the cache
	cache.ExpectEnd()
	err = writeCache(cache, cachePath)
	if err != nil {
		return errors.WithStack(err)
	}

	err = runCommand(ctx, &FailIfImportedLibraryIsWrong{})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Append the given folder to the include path and match or append it to
// the cache. sourceFilePath and include indicate the source of this
// include (e.g. what #include line in what file it was resolved from)
// and should be the empty string for the default include folders, like
// the core or variant.
func appendIncludeFolder(ctx *types.Context, cache *includeCache, sourceFilePath *paths.Path, include string, folder *paths.Path) {
	ctx.IncludeFolders = append(ctx.IncludeFolders, folder)
	cache.ExpectEntry(sourceFilePath, include, folder)
}

func runCommand(ctx *types.Context, command types.Command) error {
	PrintRingNameIfDebug(ctx, command)
	err := command.Run(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type includeCacheEntry struct {
	Sourcefile  *paths.Path
	Include     string
	Includepath *paths.Path
}

func (entry *includeCacheEntry) String() string {
	return fmt.Sprintf("SourceFile: %s; Include: %s; IncludePath: %s",
		entry.Sourcefile, entry.Include, entry.Includepath)
}

func (entry *includeCacheEntry) Equals(other *includeCacheEntry) bool {
	return entry.String() == other.String()
}

type includeCache struct {
	// Are the cache contents valid so far?
	valid bool
	// Index into entries of the next entry to be processed. Unused
	// when the cache is invalid.
	next    int
	entries []*includeCacheEntry
}

// Return the next cache entry. Should only be called when the cache is
// valid and a next entry is available (the latter can be checked with
// ExpectFile). Does not advance the cache.
func (cache *includeCache) Next() *includeCacheEntry {
	return cache.entries[cache.next]
}

// Check that the next cache entry is about the given file. If it is
// not, or no entry is available, the cache is invalidated. Does not
// advance the cache.
func (cache *includeCache) ExpectFile(sourcefile *paths.Path) {
	if cache.valid && (cache.next >= len(cache.entries) || !cache.Next().Sourcefile.EqualsTo(sourcefile)) {
		cache.valid = false
		cache.entries = cache.entries[:cache.next]
	}
}

// Check that the next entry matches the given values. If so, advance
// the cache. If not, the cache is invalidated. If the cache is
// invalidated, or was already invalid, an entry with the given values
// is appended.
func (cache *includeCache) ExpectEntry(sourcefile *paths.Path, include string, librarypath *paths.Path) {
	entry := &includeCacheEntry{Sourcefile: sourcefile, Include: include, Includepath: librarypath}
	if cache.valid {
		if cache.next < len(cache.entries) && cache.Next().Equals(entry) {
			cache.next++
		} else {
			cache.valid = false
			cache.entries = cache.entries[:cache.next]
		}
	}

	if !cache.valid {
		cache.entries = append(cache.entries, entry)
	}
}

// Check that the cache is completely consumed. If not, the cache is
// invalidated.
func (cache *includeCache) ExpectEnd() {
	if cache.valid && cache.next < len(cache.entries) {
		cache.valid = false
		cache.entries = cache.entries[:cache.next]
	}
}

// Read the cache from the given file
func readCache(path *paths.Path) *includeCache {
	bytes, err := path.ReadFile()
	if err != nil {
		// Return an empty, invalid cache
		return &includeCache{}
	}
	result := &includeCache{}
	err = json.Unmarshal(bytes, &result.entries)
	if err != nil {
		// Return an empty, invalid cache
		return &includeCache{}
	}
	result.valid = true
	return result
}

// Write the given cache to the given file if it is invalidated. If the
// cache is still valid, just update the timestamps of the file.
func writeCache(cache *includeCache, path *paths.Path) error {
	// If the cache was still valid all the way, just touch its file
	// (in case any source file changed without influencing the
	// includes). If it was invalidated, overwrite the cache with
	// the new contents.
	if cache.valid {
		path.Chtimes(time.Now(), time.Now())
	} else {
		bytes, err := json.MarshalIndent(cache.entries, "", "  ")
		if err != nil {
			return errors.WithStack(err)
		}
		err = path.WriteFile(bytes)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func findIncludesUntilDone(ctx *types.Context, cache *includeCache, sourceFile types.SourceFile) error {
	sourcePath := sourceFile.SourcePath(ctx)
	targetFilePath := paths.NullPath()
	depPath := sourceFile.DepfilePath(ctx)
	objPath := sourceFile.ObjectPath(ctx)

	// TODO: This should perhaps also compare against the
	// include.cache file timestamp. Now, it only checks if the file
	// changed after the object file was generated, but if it
	// changed between generating the cache and the object file,
	// this could show the file as unchanged when it really is
	// changed. Changing files during a build isn't really
	// supported, but any problems from it should at least be
	// resolved when doing another build, which is not currently the
	// case.
	// TODO: This reads the dependency file, but the actual building
	// does it again. Should the result be somehow cached? Perhaps
	// remove the object file if it is found to be stale?
	unchanged, err := builder_utils.ObjFileIsUpToDate(ctx, sourcePath, objPath, depPath)
	if err != nil {
		return errors.WithStack(err)
	}

	first := true
	for {
		var include string
		cache.ExpectFile(sourcePath)

		includes := ctx.IncludeFolders
		if library, ok := sourceFile.Origin.(*libraries.Library); ok && library.UtilityDir != nil {
			includes = append(includes, library.UtilityDir)
		}

		if library, ok := sourceFile.Origin.(*libraries.Library); ok {
			if library.Precompiled && library.PrecompiledWithSources {
				// Fully precompiled libraries should have no dependencies
				// to avoid ABI breakage
				if ctx.Verbose {
					ctx.GetLogger().Println(constants.LOG_LEVEL_DEBUG, constants.MSG_SKIP_PRECOMPILED_LIBRARY, library.Name)
				}
				return nil
			}
		}

		var preproc_err error
		var preproc_stderr []byte

		if unchanged && cache.valid {
			include = cache.Next().Include
			if first && ctx.Verbose {
				ctx.GetLogger().Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_CACHED_INCLUDES, sourcePath)
			}
		} else {
			preproc_stderr, preproc_err = GCCPreprocRunnerForDiscoveringIncludes(ctx, sourcePath, targetFilePath, includes)
			// Unwrap error and see if it is an ExitError.
			_, is_exit_error := errors.Cause(preproc_err).(*exec.ExitError)
			if preproc_err == nil {
				// Preprocessor successful, done
				include = ""
			} else if !is_exit_error || preproc_stderr == nil {
				// Ignore ExitErrors (e.g. gcc returning
				// non-zero status), but bail out on
				// other errors
				return errors.WithStack(preproc_err)
			} else {
				include = IncludesFinderWithRegExp(string(preproc_stderr))
				if include == "" && ctx.Verbose {
					ctx.GetLogger().Println(constants.LOG_LEVEL_DEBUG, constants.MSG_FIND_INCLUDES_FAILED, sourcePath)
				}
			}
		}

		if include == "" {
			// No missing includes found, we're done
			cache.ExpectEntry(sourcePath, "", nil)
			return nil
		}

		library := ResolveLibrary(ctx, include)
		if library == nil {
			// Library could not be resolved, show error
			// err := runCommand(ctx, &GCCPreprocRunner{SourceFilePath: sourcePath, TargetFileName: paths.New(constants.FILE_CTAGS_TARGET_FOR_GCC_MINUS_E), Includes: includes})
			// return errors.WithStack(err)
			if preproc_err == nil || preproc_stderr == nil {
				// Filename came from cache, so run preprocessor to obtain error to show
				preproc_stderr, preproc_err = GCCPreprocRunnerForDiscoveringIncludes(ctx, sourcePath, targetFilePath, includes)
				if preproc_err == nil {
					// If there is a missing #include in the cache, but running
					// gcc does not reproduce that, there is something wrong.
					// Returning an error here will cause the cache to be
					// deleted, so hopefully the next compilation will succeed.
					return errors.New("Internal error in cache")
				}
			}
			os.Stderr.Write(preproc_stderr)
			return errors.WithStack(preproc_err)
		}

		// Add this library to the list of libraries, the
		// include path and queue its source files for further
		// include scanning
		ctx.ImportedLibraries = append(ctx.ImportedLibraries, library)
		appendIncludeFolder(ctx, cache, sourcePath, include, library.SourceDir)
		sourceDirs := library.SourceDirs()
		for _, sourceDir := range sourceDirs {
			queueSourceFilesFromFolder(ctx, ctx.CollectedSourceFiles, library, sourceDir.Dir, sourceDir.Recurse)
		}
		first = false
	}
}

func queueSourceFilesFromFolder(ctx *types.Context, queue *types.UniqueSourceFileQueue, origin interface{}, folder *paths.Path, recurse bool) error {
	extensions := func(ext string) bool { return ADDITIONAL_FILE_VALID_EXTENSIONS_NO_HEADERS[ext] }

	filePaths := []string{}
	err := utils.FindFilesInFolder(&filePaths, folder.String(), extensions, recurse)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, filePath := range filePaths {
		sourceFile, err := types.MakeSourceFile(ctx, origin, paths.New(filePath))
		if err != nil {
			return errors.WithStack(err)
		}
		queue.Push(sourceFile)
	}

	return nil
}
