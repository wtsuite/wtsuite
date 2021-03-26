package main

import (
  "fmt"
  "os"

	"github.com/wtsuite/wtsuite/pkg/directives"
	"github.com/wtsuite/wtsuite/pkg/files"
	"github.com/wtsuite/wtsuite/pkg/git"
	"github.com/wtsuite/wtsuite/pkg/parsers"
	"github.com/wtsuite/wtsuite/pkg/tokens/js"
	"github.com/wtsuite/wtsuite/pkg/tokens/js/macros"
	"github.com/wtsuite/wtsuite/pkg/tokens/js/values"
	"github.com/wtsuite/wtsuite/pkg/tokens/patterns"
	"github.com/wtsuite/wtsuite/pkg/tree"
	"github.com/wtsuite/wtsuite/pkg/tree/scripts"
)

const (
  DEFAULT_OUTPUTFILE = "a.html"
)

var (
  VERSION string
  VERBOSITY = 0
  cmdParser *parsers.CLIParser = nil
)

type CmdArgs struct {
  inputFile string
  outputFile string

  control string // optional control to be built along with view
  autoDownload bool

  // stylesheets and js is included inline

  compactOutput bool
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
		control:        "",
    autoDownload: false,
		compactOutput: false,
		verbosity:     0,
	}

  cmdParser = parsers.NewCLIParser(
    fmt.Sprintf("Usage: %s <input-file> [-o <output-file>] [options]\n", os.Args[0]),
    "",
    []parsers.CLIOption{
      parsers.NewCLIVersion("", "version",   "--version    Show version", VERSION),
      parsers.NewCLIUniqueFlag("c", "compact"       , "-c, --compact          Compact output with minimal whitespace and short names", &(cmdArgs.compactOutput)),
      parsers.NewCLIUniqueFlag("", "auto-download"         , "--auto-download                   Automatically download missing packages (use wt-pkg-sync if you want to do this manually). Doesn't update packages!", &(cmdArgs.autoDownload)), 
      parsers.NewCLIUniqueFile("o", "output"        , "-o, --output <file>    Defaults to \"" + DEFAULT_OUTPUTFILE + "\" if not set", false, &(cmdArgs.outputFile)),
      parsers.NewCLIUniqueFile("", "control"        , "--control <file>       Optional control file", true, &(cmdArgs.control)),
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
  if cmdArgs.compactOutput {
		tree.COMPRESS_NUMBERS = true
		patterns.NL = ""
		patterns.TAB = ""
		patterns.LAST_SEMICOLON = ""
    patterns.COMPACT_NAMING = true
    macros.COMPACT = true
  }

  if cmdArgs.autoDownload {
    git.RegisterFetchPublicOrPrivate()
  }

	directives.ForceNewViewFileScriptRegistration(directives.NewFileCache())

  js.TARGET = "browser"

	VERBOSITY = cmdArgs.verbosity
	files.VERBOSITY = cmdArgs.verbosity
	parsers.VERBOSITY = cmdArgs.verbosity
	js.VERBOSITY = cmdArgs.verbosity
	values.VERBOSITY = cmdArgs.verbosity
	scripts.VERBOSITY = cmdArgs.verbosity

  return files.ResolvePackages(cmdArgs.inputFile)
}

func buildHTMLFile(c *directives.FileCache, src string, dst string, control string, compactOutput bool) error {
  r, err := directives.NewRoot(c, src)
  if err != nil {
    return err
  }

  if control != "" {
    entryScript, err := scripts.NewInitFileScript(control)
    if err != nil {
      return err
    }

		bundle := scripts.NewFileBundle(map[string]string{})

		bundle.Append(entryScript)

		if err := bundle.Finalize(); err != nil {
			return err
		}

		content, err := bundle.Write()
		if err != nil {
			return err
		}

    if err := r.IncludeScript(content); err != nil {
      return err
    }
  }

	output := r.Write("", patterns.NL, patterns.TAB)

	// src is just for info
	if err := files.WriteFile(src, dst, []byte(output)); err != nil {
		return err
	}

  return nil
}

func buildFile(cmdArgs CmdArgs) error {
  c := directives.NewFileCache()

	if err := buildHTMLFile(c, cmdArgs.inputFile, cmdArgs.outputFile, cmdArgs.control, cmdArgs.compactOutput); err != nil {
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
