package scaffold

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"
)

// Exit codes for the scaffold command.
const (
	exitOK      = 0 // project generated
	exitFailure = 1 // generation, tidy, or git step failed
	exitUsage   = 2 // bad or missing flags
)

// Run parses args, builds a Spec, generates the project, and runs the
// post-generation steps (`go mod tidy`, `git init`) unless skipped. It writes
// progress to stdout and errors to stderr, returning a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("scaffold", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		source   = fs.String("source", "template", "path to the template/ directory to copy from")
		dest     = fs.String("dest", "", "destination directory for the new project (required)")
		module   = fs.String("module", "", "new module path, e.g. github.com/you/tool (required)")
		name     = fs.String("name", "", "tool/binary name, e.g. tool (required)")
		author   = fs.String("author", "", "copyright holder for community files")
		year     = fs.String("year", "", "copyright year (defaults to the current year)")
		desc     = fs.String("description", "", "one-line tool description")
		email    = fs.String("email", "", "contact/maintainer email for community + release files")
		skipTidy = fs.Bool("skip-tidy", false, "do not run `go mod tidy` in the generated project")
		skipGit  = fs.Bool("skip-git", false, "do not run `git init` in the generated project")
	)

	if err := fs.Parse(args); err != nil {
		return exitUsage // flag package already printed the error + usage
	}

	var missing []string
	if *module == "" {
		missing = append(missing, "-module")
	}
	if *name == "" {
		missing = append(missing, "-name")
	}
	if *dest == "" {
		missing = append(missing, "-dest")
	}
	if len(missing) > 0 {
		fmt.Fprintf(stderr, "scaffold: missing required flag(s): %v\n", missing)
		fs.Usage()
		return exitUsage
	}

	resolvedYear := *year
	if resolvedYear == "" {
		resolvedYear = strconv.Itoa(time.Now().Year())
	}

	spec := Spec{
		Source:      *source,
		Dest:        *dest,
		Module:      *module,
		Name:        *name,
		Author:      *author,
		Year:        resolvedYear,
		Description: *desc,
		Email:       *email,
	}

	if err := Generate(spec); err != nil {
		fmt.Fprintf(stderr, "scaffold: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "Generated %s in %s\n", spec.Name, spec.Dest)

	if !*skipTidy {
		fmt.Fprintln(stdout, "Running go mod tidy...")
		if err := runCmd(spec.Dest, stdout, stderr, "go", "mod", "tidy"); err != nil {
			fmt.Fprintf(stderr, "scaffold: go mod tidy: %v\n", err)
			return exitFailure
		}
	}
	if !*skipGit {
		fmt.Fprintln(stdout, "Initializing git repository...")
		if err := runCmd(spec.Dest, stdout, stderr, "git", "init", "-q"); err != nil {
			fmt.Fprintf(stderr, "scaffold: git init: %v\n", err)
			return exitFailure
		}
	}

	fmt.Fprintf(stdout, "Done. cd %s\n", spec.Dest)
	return exitOK
}

// runCmd runs name+args in dir, streaming output to the provided writers.
func runCmd(dir string, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
