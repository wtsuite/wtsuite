package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/computeportal/wtsuite/pkg/directives"
	"github.com/computeportal/wtsuite/pkg/files"
	"github.com/computeportal/wtsuite/pkg/git"
	"github.com/computeportal/wtsuite/pkg/parsers"
	"github.com/computeportal/wtsuite/pkg/tokens/js"
	"github.com/computeportal/wtsuite/pkg/tokens/js/macros"
	"github.com/computeportal/wtsuite/pkg/tokens/js/values"
	"github.com/computeportal/wtsuite/pkg/tokens/patterns"
	"github.com/computeportal/wtsuite/pkg/tree/scripts"
)

const (
	DEFAULT_OUTPUTFILE = "a.js"
	DEFAULT_TARGET     = "nodejs"
)

var (
  VERSION string
	VERBOSITY = 0
  cmdParser *parsers.CLIParser
)

type CmdArgs struct {
	inputFile   string // entry script
	outputFile  string // defaults to a.js in current dir
	target      string

	compactOutput bool
	forceBuild    bool // delete cache and start fresh
  executable    bool // create an executable
  autoDownload  bool

  globalVars map[string]string

	verbosity int
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "%s\n", cmdParser.Info())
	os.Exit(1)
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
		target:        DEFAULT_TARGET,
		compactOutput: false,
		forceBuild:    false,
    executable:    false,
    globalVars:    make(map[string]string),
    autoDownload:  false,
		verbosity:     0,
	}

  cmdParser = parsers.NewCLIParser(
    fmt.Sprintf("Usage: %s <input-file> [-o <output-file>] [options]\n", os.Args[0]),
    "",
    []parsers.CLIOption{
      parsers.NewCLIVersion("", "version",   "--version    Show version", VERSION),
      parsers.NewCLIUniqueFile("o", "output"    ,   "-o, --output <output-file>  Defaults to \"" + DEFAULT_OUTPUTFILE + "\" if not set", false, &(cmdArgs.outputFile)),
      parsers.NewCLIUniqueFlag("c", "compact"   ,   "-c, --compact               Compact output with minimal whitespace and short names", &(cmdArgs.compactOutput)),
      parsers.NewCLIUniqueFlag("f", "force"     ,   "-f, --force                 Force a complete project rebuild", &(cmdArgs.forceBuild)),
      parsers.NewCLIUniqueEnum("t", "target"    ,   "-t, --target <js-target>    Defaults to \"" + DEFAULT_TARGET + "\", other possibilities are \"browser\" or \"worker\"", []string{"nodejs", "browser", "worker"}, &(cmdArgs.target)),
      parsers.NewCLIUniqueFlag("x", "executable",   "-x, --executable            Create an executable with a node hashbang (target must be nodejs)", &(cmdArgs.executable)),
      parsers.NewCLIUniqueKeyValue("D"                 , "-D<name> <value>              Define a global variable with a string value", cmdArgs.globalVars),
      parsers.NewCLIUniqueKey("B"                      , "-B<name>                      Define a global flag (its value is an empty string)", cmdArgs.globalVars),
      parsers.NewCLIUniqueFlag("", "auto-download", "--auto-download             Automatically download missing packages (use wt-pkg-sync if you want to do this manually). Doesn't update packages!", &(cmdArgs.autoDownload)), 
      parsers.NewCLIUniqueFlag("l", "latest"    ,   "-l, --latest                Ignore max semver, use latest tagged versions of dependencies", &(files.LATEST)),
      parsers.NewCLICountFlag("v", ""           ,   "-v[v[v..]]                  Verbosity", &(cmdArgs.verbosity)),
    },
    parsers.NewCLIFile("", "", "", true, &(cmdArgs.inputFile)),
  )

  if err := cmdParser.Parse(os.Args[1:]); err != nil {
    printMessageAndExit(err.Error())
  }

  if cmdArgs.executable && cmdArgs.target != DEFAULT_TARGET {
    printMessageAndExit("Error: --executable can only be used if target is nodejs")
  }

	return cmdArgs
}

func setUpEnv(cmdArgs CmdArgs) error {
	if cmdArgs.compactOutput {
    patterns.NL = ""
		patterns.TAB = ""
		patterns.COMPACT_NAMING = true
		macros.COMPACT = true
	}

  if cmdArgs.autoDownload {
    git.RegisterFetchPublicOrPrivate()
  }

	js.TARGET = cmdArgs.target
	directives.ForceNewViewFileScriptRegistration(directives.NewFileCache())

	VERBOSITY = cmdArgs.verbosity
	files.VERBOSITY = cmdArgs.verbosity
	parsers.VERBOSITY = cmdArgs.verbosity
	js.VERBOSITY = cmdArgs.verbosity
	values.VERBOSITY = cmdArgs.verbosity
	scripts.VERBOSITY = cmdArgs.verbosity

  return files.ResolvePackages(cmdArgs.inputFile)
}

func buildProject(cmdArgs CmdArgs) error {
  files.LoadDepTree(cmdArgs.outputFile, "", cmdArgs.forceBuild)

  defer files.SaveDepTree()

	if files.RequiresDepUpdate(cmdArgs.outputFile, "") {
    files.StartDstUpdate(cmdArgs.outputFile, "")
    files.AddDep(cmdArgs.outputFile, cmdArgs.inputFile)
    
		entryScript, err := scripts.NewInitFileScript(cmdArgs.inputFile)
		if err != nil {
			return err
		}

		bundle := scripts.NewFileBundle(cmdArgs.globalVars)

		bundle.Append(entryScript)

		if err := bundle.Finalize(); err != nil {
			return err
		}

		content, err := bundle.Write()
		if err != nil {
			return err
		}

    if cmdArgs.executable {
      if err := ioutil.WriteFile(cmdArgs.outputFile, []byte("#!/usr/bin/env node\n"+content), 0755); err != nil {
        return errors.New("Error: " + err.Error())
      }
    } else {
      if err := ioutil.WriteFile(cmdArgs.outputFile, []byte(content), 0644); err != nil {
        return errors.New("Error: " + err.Error())
      }
    }
	}

	return nil
}

func main() {
	cmdArgs := parseArgs()

  if err := setUpEnv(cmdArgs); err != nil {
		printSyntaxErrorAndExit(err)
  }

	if err := buildProject(cmdArgs); err != nil {
		printSyntaxErrorAndExit(err)
	}
}
