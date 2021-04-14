package main

import (
  "errors"
  "fmt"
  "io/ioutil"
  "os"
  "os/signal"
  "path/filepath"
  "runtime/pprof"
  "sort"
  "strings"

	"github.com/wtsuite/wtsuite/pkg/directives"
	"github.com/wtsuite/wtsuite/pkg/files"
	"github.com/wtsuite/wtsuite/pkg/git"
	"github.com/wtsuite/wtsuite/pkg/parsers"
	"github.com/wtsuite/wtsuite/pkg/styles"
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
	"github.com/wtsuite/wtsuite/pkg/tokens/js"
	"github.com/wtsuite/wtsuite/pkg/tokens/js/macros"
	"github.com/wtsuite/wtsuite/pkg/tokens/js/values"
	"github.com/wtsuite/wtsuite/pkg/tokens/patterns"
	"github.com/wtsuite/wtsuite/pkg/tree"
	"github.com/wtsuite/wtsuite/pkg/tree/scripts"
)

var (
  VERSION string
  VERBOSITY = 0
  cmdParser *parsers.CLIParser = nil
)

type CmdArgs struct {
  configFile     string
  outputDir      string
  globals        map[string]string

  compactOutput  bool
  forceRebuild   bool
  autoDownload   bool
  clean          bool

  profFile       string
  verbosity      int
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
    configFile:    "site-config.thtml",
    outputDir:     "./www",
    globals:       make(map[string]string),
    compactOutput: false,
    forceRebuild:  false,
    autoDownload:  false,
    clean:         false,
    profFile:      "",
    verbosity:     0,
  }

  var positional []string = nil

  cmdParser = parsers.NewCLIParser(
    fmt.Sprintf("Usage: %s [options]\n", os.Args[0]),
    "",
    []parsers.CLIOption{
      parsers.NewCLIVersion("", "version",   "--version    Show version", VERSION),
      parsers.NewCLIUniqueFile("i", "input"              , "-i, --input <config-file>                  Config file", true, &(cmdArgs.configFile)),
      parsers.NewCLIUniqueFlag("c", "compact",      "-c, --compact    Minified output", &(cmdArgs.compactOutput)),
      parsers.NewCLIUniqueFile("o", "output"              , "-o, --output <output-dir>                  Defaults to ./www", false, &(cmdArgs.outputDir)),
      parsers.NewCLIUniqueFlag("f", "force",        "-f, --force      Force a complete build", &(cmdArgs.forceRebuild)),
      parsers.NewCLIUniqueFlag("", "auto-download", "--auto-download  Automatically download missing packages. Doesnt update!", &(cmdArgs.autoDownload)),
      parsers.NewCLIUniqueFlag("", "clean", "--clean  Delete files in dst directory that are not a result of this build", &(cmdArgs.clean)),
      parsers.NewCLIUniqueFlag("l", "latest"           , "-l, --latest                  Ignore max semver, use latest tagged versions of dependencies", &(files.LATEST)),
      parsers.NewCLICountFlag("v" , ""                 , "-v[v[v..]]                    Verbosity", &(cmdArgs.verbosity)),
      parsers.NewCLIUniqueKeyValue("D"                 , "-D<name> <value>              Define a global variable with a string value", cmdArgs.globals),
      parsers.NewCLIUniqueKey("B"                      , "-B<name>                      Define a global flag (its value is an empty string)", cmdArgs.globals),
      parsers.NewCLIUniqueFile("", "prof"              , "--prof<file>                  Profile the transpiler, output written to file (analyzeable with go tool pprof)", false, &(cmdArgs.profFile)),
    },
    parsers.NewCLIRemaining(&positional),
  )

  if err := cmdParser.Parse(os.Args[1:]); err != nil {
    printMessageAndExit(err.Error())
  }

  if len(positional) != 0 {
    printMessageAndExit("Error: unexpected positional arguments")
  }

  if !filepath.IsAbs(cmdArgs.configFile) {
    absConfigFile, err := filepath.Abs(cmdArgs.configFile)
    if err != nil {
      printMessageAndExit(err.Error())
    }

    cmdArgs.configFile = absConfigFile
  }

  if !files.IsFile(cmdArgs.configFile) {
    printMessageAndExit("Error: config file \"" + cmdArgs.configFile + "\" not found")
  }

  if !filepath.IsAbs(cmdArgs.outputDir) {
    absOutputDir, err := filepath.Abs(cmdArgs.outputDir)
    if err != nil {
      printMessageAndExit(err.Error())
    }

    cmdArgs.outputDir = absOutputDir
  }

  pwd, err := os.Getwd()
  if err != nil {
    printMessageAndExit(err.Error())
  }

  if cmdArgs.outputDir == pwd {
    // TODO: outputDir must be empty?
    printMessageAndExit("Error: output dir can't be same as current dir")
  }

  if err := os.MkdirAll(cmdArgs.outputDir, 0755); err != nil {
    printMessageAndExit(err.Error())
  }

  return cmdArgs
}

func setUpEnv(cmdArgs CmdArgs, cfg *SiteConfig) error {
	if cmdArgs.compactOutput {
		patterns.NL = ""
		patterns.TAB = ""
		patterns.LAST_SEMICOLON = ""
    patterns.COMPACT_NAMING = true
    macros.COMPACT = true
		tree.COMPRESS_NUMBERS = true
	}

  if cmdArgs.autoDownload {
    git.RegisterFetchPublicOrPrivate()
  }

  lst := make([]string, 0)
	for k, v := range cmdArgs.globals {
    lst = append(lst, k)
		directives.RegisterDefine(k, v)
	}
  sort.Strings(lst)
  var b strings.Builder
  b.WriteString("globals:{")
  for i, k := range lst {
    b.WriteString(k)
    b.WriteString(":")
    b.WriteString(cmdArgs.globals[k])
    if i < len(lst) - 1 {
      b.WriteString(";")
    }
  }
  b.WriteString(",version:")
  b.WriteString(VERSION)

	directives.ForceNewViewFileScriptRegistration(directives.NewFileCache())

	VERBOSITY = cmdArgs.verbosity
	directives.VERBOSITY = cmdArgs.verbosity
	tokens.VERBOSITY = cmdArgs.verbosity
	js.VERBOSITY = cmdArgs.verbosity
	values.VERBOSITY = cmdArgs.verbosity
	parsers.VERBOSITY = cmdArgs.verbosity
	files.VERBOSITY = cmdArgs.verbosity
	tree.VERBOSITY = cmdArgs.verbosity
	scripts.VERBOSITY = cmdArgs.verbosity

  files.LoadDepTree(cmdArgs.outputDir, b.String(), cmdArgs.forceRebuild)

  directives.MATH_FONT_URL = cfg.MathFontURL()
  styles.SaveMathFont(cfg.MathFontDst())

  return nil
}

var fProf *os.File = nil

func startProfiling(profFile string) {
  var err error
  fProf, err = os.Create(profFile)
  if err != nil {
    printMessageAndExit(err.Error())
  }

  pprof.StartCPUProfile(fProf)

  go func() {
    sigchan := make(chan os.Signal)
    signal.Notify(sigchan, os.Interrupt)
    <-sigchan

    stopProfiling(profFile)

    os.Exit(1)
  }()
}

func stopProfiling(profFile string) {
  if fProf != nil {
		pprof.StopCPUProfile()

    // also write mem profile
		fMem, err := os.Create(profFile + ".mprof")
		if err != nil {
			printMessageAndExit(err.Error())
		}

		pprof.WriteHeapProfile(fMem)
		fMem.Close()

    fProf = nil
  }
}

func copyFile(src, dst string) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	// src is just for info
	if err := files.WriteFile(src, dst, content); err != nil {
		return err
	}

	return nil
}

func buildSiteFiles(cfg *SiteConfig, cmdArgs CmdArgs) error {
  for _, f := range cfg.Files {
    if files.RequiresDepUpdate(f.dst, "") {
      files.StartDstUpdate(f.dst, "")
      if err := copyFile(f.src, f.dst); err != nil {
        return err
      }
    }
  }

  return nil
}

func cleanURL(url string) string {
  if !strings.HasPrefix(url, "/") {
    return "/" + url
  } else {
    return url
  }
}

func cleanLink(this string, link string) string {
  this = cleanURL(this)[1:]

  if len(strings.Split(this, "/")) > 1 {
    return cleanURL(link)
  } else {
    return cleanURL(link)[1:]
  }
}

func buildSiteStyles(cfg *SiteConfig, cmdArgs CmdArgs) error {
  for i, style := range cfg.Styles {
    // if any of the pages requires an update, then the sheet definitely needs an update
    somePagesNeedUpdate := false
    for _, pURL := range style.pages {
      p := cfg.FindPage(pURL)
      if files.RequiresDepUpdate(p.dst, cfg.PageParameterString(pURL)) {
        somePagesNeedUpdate = true
        break
      }
    }

    if files.RequiresDepUpdate(style.dst, "") || somePagesNeedUpdate {
      files.StartDstUpdate(style.dst, "")
      files.AddDep(style.dst, style.src)

      sheet, err := styles.Build(style.src, context.NewDummyContext())
      if err != nil {
        return err
      }

      if err = styles.WriteSheetToFile(sheet, style.dst); err != nil {
        return err
      }

      // save back in config for use by pages
      style.sheet = sheet
      cfg.Styles[i] = style
    }
  }

  return nil
}

func buildSitePages(cfg *SiteConfig, cmdArgs CmdArgs) error {
  cache := directives.NewFileCache()

  for _, page := range cfg.Pages {
    if len(page.params) == 0 {
      directives.RegisterURL(page.src, page.url)
    }
  }

  for _, page := range cfg.Pages {
    parameters := cfg.PageParameterString(page.url)

    if files.RequiresDepUpdate(page.dst, parameters) {
      files.StartDstUpdate(page.dst, parameters)
      files.AddDep(page.dst, page.src)

      directives.SetActiveURL(cleanURL(page.url))

      r, err := directives.NewRoot(cache, page.src)
      if err != nil {
        return err
      }
      directives.UnsetActiveURL()

      for _, styleURL := range cfg.PageStyles(page.url) {
        r.LinkStyle(cleanLink(page.url, styleURL))

        s := cfg.FindStyle(styleURL)
        files.AddDep(page.dst, s.src)

        r, err = s.sheet.ApplyExtensions(r)
        if err != nil {
          return err
        }
      }

      scriptHashes := cfg.PageScripts(page.url)
      if len(scriptHashes) > 0 {
        r.LinkScriptBundle(cleanLink(page.url, cfg.JSURL()), scriptHashes)
      }

      output := r.Write("", patterns.NL, patterns.TAB)

      if err := files.WriteFile(page.src, page.dst, []byte(output)); err != nil {
        return err
      }
    }
  }

  return nil
}

func buildSiteScripts(cfg *SiteConfig, cmdArgs CmdArgs) error {
  dst := cfg.JSDst()

  if files.RequiresDepUpdate(dst, "") {
    files.StartDstUpdate(dst, "")

    js.TARGET = "browser"

    bundle := scripts.NewFileBundle(cmdArgs.globals)

    for _, script := range cfg.Scripts {
      files.AddDep(dst, script.src)

      sc, err := scripts.NewControlFileScript(script.src, script.hash)
      if err != nil {
        return err
      }

      bundle.Append(sc)
    }

    if err := bundle.Finalize(); err != nil {
      return err
    }

		content, err := bundle.Write()
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(dst, []byte(content), 0644); err != nil {
			return errors.New("Error: " + err.Error())
		}
  }

  return nil
}

func buildSite(cmdArgs CmdArgs, cfg *SiteConfig) error {
  defer files.SaveDepTree()

	if err := buildSiteFiles(cfg, cmdArgs); err != nil {
		return err
	}

	if err := buildSiteStyles(cfg, cmdArgs); err != nil {
		return err
	}

	if err := buildSitePages(cfg, cmdArgs); err != nil {
		return err
	}

	if err := buildSiteScripts(cfg, cmdArgs); err != nil {
		return err
	}

  if cmdArgs.clean {
    if err := cfg.CleanOutput(); err != nil {
      return err
    }
  }

	return nil
}

func main() {
  cmdArgs := parseArgs()

  cfg, err := ReadConfigFile(cmdArgs.configFile, cmdArgs.outputDir)
  if err != nil {
    printMessageAndExit(err.Error())
  }

  if err := setUpEnv(cmdArgs, cfg); err != nil {
		printMessageAndExit(err.Error())
  }

	if cmdArgs.profFile != "" {
    startProfiling(cmdArgs.profFile)
	}

	if err := buildSite(cmdArgs, cfg); err != nil {
		printSyntaxErrorAndExit(err)
	}

	if cmdArgs.profFile != "" {
    stopProfiling(cmdArgs.profFile)
	}
}
