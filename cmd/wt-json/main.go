package main

import (
  "fmt"
  "os"

	"github.com/wtsuite/wtsuite/pkg/directives"
	"github.com/wtsuite/wtsuite/pkg/files"
	"github.com/wtsuite/wtsuite/pkg/git"
	"github.com/wtsuite/wtsuite/pkg/parsers"
	_ "github.com/wtsuite/wtsuite/pkg/styles" // for side-effect
)

const (
  DEFAULT_OUTPUTFILE = "a.json"
)

var (
  VERSION string
  VERBOSITY = 0
  cmdParser *parsers.CLIParser = nil
)

type CmdArgs struct {
  inputFile string
  outputFile string

  autoDownload bool
  verbosity int
}

func printMessageAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "\u001b[1m"+msg+"\u001b[0m\n\n")
  os.Exit(1)
}

func printSyntaxErrorAndExit(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

func parseArgs() CmdArgs {
	cmdArgs := CmdArgs{
		inputFile:     "",
		outputFile:    DEFAULT_OUTPUTFILE,
    autoDownload:  false,
		verbosity:     0,
	}

  cmdParser = parsers.NewCLIParser(
    fmt.Sprintf("Usage: %s <input-file> [-o <output-file>] [options]\n", os.Args[0]),
    "",
    []parsers.CLIOption{
      parsers.NewCLIVersion("", "version",   "--version    Show version", VERSION),
      parsers.NewCLIUniqueFile("o", "output"        , "-o, --output <file>    Defaults to \"" + DEFAULT_OUTPUTFILE + "\" if not set", false, &(cmdArgs.outputFile)),
      parsers.NewCLIUniqueFlag("", "auto-download"         , "--auto-download                   Automatically download missing packages (use wt-pkg-sync if you want to do this manually). Doesn't update packages!", &(cmdArgs.autoDownload)), 
      parsers.NewCLIUniqueFlag("l", "latest"        , "-l, --latest           Ignore max semver, use latest tagged versions of dependencies", &(files.LATEST)),
      parsers.NewCLICountFlag("v", ""               , "-v[v[v..]]             Verbosity", &(cmdArgs.verbosity)),
    },
    parsers.NewCLIFile("", "", "", true, &(cmdArgs.inputFile)),
  )

  if err := cmdParser.Parse(os.Args[1:]); err != nil {
    printMessageAndExit(err.Error())
  }

  return cmdArgs
}

func setUpEnv(cmdArgs CmdArgs) error {
  if cmdArgs.autoDownload {
    git.RegisterFetchPublicOrPrivate()
  }

	VERBOSITY = cmdArgs.verbosity
	files.VERBOSITY = cmdArgs.verbosity
	parsers.VERBOSITY = cmdArgs.verbosity

  return files.ResolvePackages(cmdArgs.inputFile)
}

func buildFile(cmdArgs CmdArgs) error {
  if err := directives.BuildJSONFile(cmdArgs.inputFile, cmdArgs.outputFile); err != nil {
    return err
  }

  return nil
}

func main() {
  cmdArgs := parseArgs()

  if err := setUpEnv(cmdArgs); err != nil {
    printSyntaxErrorAndExit(err)
  }

  if err := buildFile(cmdArgs); err != nil {
    printSyntaxErrorAndExit(err)
  }
}
