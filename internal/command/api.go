package command

type CommandInfo interface {
	Name() string

	Type() string

	Category() string

	Group() string

	ArgsUsage() string

	Examples() []ExampleEntry

	ShortDescription() string

	LongDescription() string

	DocFile() string

	DocLink() string
}

type CommandManifest interface {
	CommandInfo

	Executable() string

	Arguments() []string

	ValidArgs() []string

	ValidArgsCmd() []string

	RequiredFlags() []string

	FlagValuesCmd() []string

	CheckFlags() bool
}

type Command interface {
	CommandManifest

	Execute(envVars []string, args ...string) (int, error)

	ExecuteValidArgsCmd(envVars []string, args ...string) (int, string, error)

	ExecuteFlagValuesCmd(envVars []string, args ...string) (int, string, error)

	SetPkgDir(pkgDir string)
}

type PackageManifest interface {
	Name() string

	Version() string

	Commands() []Command
}

type Package interface {
	PackageManifest

	InstallTo(pathname string) (PackageManifest, error)
}

type ExampleEntry struct {
	Scenario string `json:"scenario" yaml:"scenario"`
	Command  string `json:"cmd" yaml:"cmd"`
}

func (example ExampleEntry) Clone() ExampleEntry {
	return ExampleEntry{
		Scenario: example.Scenario,
		Command:  example.Command,
	}
}
