package types

import (
	"github.com/arduino/arduino-cli/arduino/libraries"
)

type CodeModelGCCInvocation struct {
	GCC        string
	InputFile  string
	ObjectFile string
	Arguments  []string
}

type CodeModelLibrary struct {
	Name            string
	SourceDirectory string
	ArchiveFile     string
	Invocations     []*CodeModelGCCInvocation
}

type KnownLibrary struct {
	Folder        string
	SrcFolder     string
	UtilityFolder string
	Layout        libraries.LibraryLayout
	Name          string
	RealName      string
	IsLegacy      bool
	Version       string
	Author        string
	Maintainer    string
	Sentence      string
	Paragraph     string
	URL           string
	Category      string
	License       string
}

type KnownHeader struct {
	Name               string
	LibraryDirectories []string
}

type KeyValuePair struct {
	Key   string
	Value string
}

type CodeModelBuilder struct {
	Core              *CodeModelLibrary
	Sketch            *CodeModelLibrary
	Libraries         []*CodeModelLibrary
	KnownHeaders      []*KnownHeader
	Prototypes        []*Prototype
	KnownLibraries    []*KnownLibrary
	LinkerCommandLine string
	BuildProperties   []KeyValuePair
}
