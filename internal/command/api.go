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

	RequestedResources() []string
}

type CommandManifest interface {
	CommandInfo

	Executable() string

	Arguments() []string

	ValidArgs() []string

	ValidArgsCmd() []string

	// deprecated in 1.9.0, replaced by Flags()
	RequiredFlags() []string

	Flags() []Flag

	ExclusiveFlags() [][]string

	GroupFlags() [][]string

	FlagValuesCmd() []string

	CheckFlags() bool
}

type Command interface {
	CommandManifest

	// the id of the reigstry that the command belongs to
	RepositoryID() string
	// the package name that the command belongs to
	PackageName() string
	// the full ID of the command: registry:package:group:name
	ID() string
	// the full group name: registry:package:group
	FullGroup() string
	// the full command name: registry:package:group:name
	FullName() string
	// the runtime group of the command
	RuntimeGroup() string
	// the runtime name of the command
	RuntimeName() string
	// the package directory
	PackageDir() string

	Execute(envVars []string, args ...string) (int, error)

	ExecuteWithOutput(envVars []string, args ...string) (int, string, error)

	ExecuteValidArgsCmd(envVars []string, args ...string) (int, string, error)

	ExecuteFlagValuesCmd(envVars []string, flagCmd []string, args ...string) (int, string, error)

	// namespace speficies the package and the registry/repository of the command
	// there could be two commands with the same group and name in different namespace
	// when resolving the group and name conflict, namespace is used to identify the
	// command
	SetNamespace(regId string, pkgName string)

	SetPackageDir(pkgDir string)

	SetRuntimeGroup(alias string)

	SetRuntimeName(alias string)
}

type PackageManifest interface {
	Name() string

	Version() string

	Commands() []Command
}

type Package interface {
	PackageManifest

	// repository ID: dropin, default, repo1, repo2, ...
	RepositoryID() string

	// verify the sha256 checksum
	VerifyChecksum(checksum string) (bool, error)

	// verify the package signature
	VerifySignature(signature string) (bool, error)

	// install package to a local repository
	InstallTo(pathname string) (PackageManifest, error)

	// run setup process of the package
	RunSetup(pkgDir string) error
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

type Flag struct {
	FlagName        string   `json:"name" yaml:"name"`
	FlagType        string   `json:"type" yaml:"type"`
	FlagShortName   string   `json:"short" yaml:"short"`
	FlagDescription string   `json:"desc" yaml:"desc"`
	FlagDefault     string   `json:"default" yaml:"default"`
	FlagRequired    bool     `json:"required" yaml:"required"`
	FlagValues      []string `json:"values" yaml:"values"`
	FlagValuesCmd   []string `json:"valuesCmd" yaml:"valuesCmd"`
}

func (f Flag) Name() string {
	return f.FlagName
}

func (f Flag) Type() string {
	if f.FlagType != "string" && f.FlagType != "bool" {
		return "string"
	}
	return f.FlagType
}

func (f Flag) ShortName() string {
	return f.FlagShortName
}

func (f Flag) Description() string {
	return f.FlagDescription
}

func (f Flag) Required() bool {
	return f.FlagRequired
}

func (f Flag) Default() string {
	if f.FlagType == "bool" {
		return "false"
	}
	return f.FlagDefault
}

func (f Flag) Values() []string {
	if f.FlagValues != nil && len(f.FlagValues) > 0 {
		return f.FlagValues
	}
	return []string{}
}

func (f Flag) ValuesCmd() []string {
	if f.FlagValuesCmd != nil && len(f.FlagValuesCmd) > 0 {
		return f.FlagValuesCmd
	}
	return []string{}
}
