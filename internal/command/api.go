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

	RequiredFlags() []string

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

	ExecuteFlagValuesCmd(envVars []string, args ...string) (int, string, error)

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
