package main

import (
  "bufio"
  "fmt"
  "io"
  //"io/ioutil"
  "os"
  "os/exec"
  "path/filepath"
  "sort"
  "strings"

  "golang.org/x/term"
	"github.com/computeportal/wtsuite/pkg/directives"
	"github.com/computeportal/wtsuite/pkg/parsers"
	"github.com/computeportal/wtsuite/pkg/terminal"
)

var cmdParser *parsers.CLIParser = nil

type CmdArgs struct {
  quiet bool
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
  cmdArgs := CmdArgs{}

  var positional []string = nil

  cmdParser = parsers.NewCLIParser(
    fmt.Sprintf("Usage: %s [-?|-h|--help]\n", os.Args[0]),
    "Note: test the thtml expressions you would use inside function statements",
    []parsers.CLIOption{
      parsers.NewCLIUniqueFlag("q", "", "-q     Quiet, don't display any messages", &(cmdArgs.quiet)),
    },
    parsers.NewCLIRemaining(&positional),
  )

  if len(positional) != 0 {
    printMessageAndExit("Error: unexpected positional args")
  }

  return cmdArgs
}

//func completer(d prompt.Document) []prompt.Suggest {
  //return []prompt.Suggest{}
//}

func readLine(reader *bufio.Reader) string {
  line := make([]byte, 0)
  for true {
    c, err := reader.ReadByte() 
    if err != nil {
      return ""
    }

    if c == '\n' {
      return string(line)
    } else {
      fmt.Println("got ", c)
      line = append(line, c)
    }
  }

  return ""
}

var helpTopics map[string]string = map[string]string{
  "":         
`?keyboard  keyboard shortcuts in wt-repl
?vars      list all variables`,

  "keyboard": `CTRL-C            clear line
CTRL-L            clear screen
CTRL-P/ArrowUp    move backward in command history
CTRL-N/ArrowDown  move forward in command history
CTRL-D            quit`,
}

func printTermHelp(t *term.Terminal, topic string) {
  msg, ok := helpTopics[topic]
  if ok {
    fmt.Fprintln(t, msg)
  } else {
    fmt.Fprintf(t, "Unknown topic \"%s\"\n", topic)
  }
}

type StdinIntercept struct {
  ignoreLine bool
  fnClose func()
}

func (fd *StdinIntercept) Read(p []byte) (int, error) {
  n, err := os.Stdin.Read(p)
  if err != nil {
    return n, err
  }

  switch p[0] {
  case 3:
    fd.ignoreLine = true
    p[0] = 13
  case 4:
    fd.fnClose()
    fmt.Println()
    os.Exit(0)
  default:
    fmt.Println("intercepted: ", p[0:10])
  }
    
  return n, err
}

func parseExpression(scope directives.Scope, t *term.Terminal, exprStr string) {
  // create some tokens
  pwd, err := os.Getwd()
  if err != nil {
    fmt.Fprintln(t, err.Error())
    return
  }

  p, err := parsers.NewTemplateParser(exprStr, filepath.Join(pwd, "(stdin)"))
  if err != nil {
    fmt.Fprintln(t, err.Error())
    return
  }

  expr, err := p.BuildSingleExpression()
  if err != nil {
    fmt.Fprintln(t, err.Error())
    return
  }

  res, err := expr.Eval(scope)
  if err != nil {
    fmt.Fprintln(t, err.Error())
    return
  }

  fmt.Fprintln(t, res.Dump(""))
  return
}

func listAllVars(scope directives.Scope, t *term.Terminal) {
  lst := scope.ListValidVarNames()

  sort.Strings(lst)
  for _, n := range lst {
    fmt.Fprintln(t, n)
  }
}

func termLoop(cmdArgs CmdArgs) error {
  oldState, err := term.MakeRaw(0)
  if err != nil {
    return err
  }

  fnClose := func() {
    term.Restore(0, oldState)
  }

  defer fnClose()
  
  input := &StdinIntercept{
    false,
    fnClose,
  }

  screen := struct{
    io.Reader 
    io.Writer
  }{input, os.Stdout}

  t := term.NewTerminal(screen, "")
  t.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
    // TODO
    return "", 0, false
  }

  t.SetPrompt("> ")

  // cache in case we call import (not yet possible though)
  cache := directives.NewFileCache()
  scope := directives.NewFileScope(true, cache)

  for {
    line, err := t.ReadLine() 
    if err == io.EOF {
      return nil
    }

    if err != nil {
      return err
    }

    if line == "" || input.ignoreLine {
      input.ignoreLine = false
      continue
    }

    line = strings.TrimSpace(line)

    if line == "" {
      fmt.Fprintln(t, line)
    } else if strings.HasPrefix(line, "?") {
      topic := strings.TrimSpace(strings.TrimPrefix(line, "?"))

      if topic == "vars" {
        listAllVars(scope, t)
      } else {
        printTermHelp(t, topic)
      }
    } else {
      expressions := strings.Split(line, ";")

      nonEmpty := false
      for _, expression := range expressions {
        if expression != "" {
          parseExpression(scope, t, expression)
          nonEmpty = true
        }
      }

      if !nonEmpty {
        fmt.Fprintln(t, "")
      }
    }
  }
}

type Handler struct {
  t *terminal.Terminal
}

type StdoutIntercept struct {
  b []byte
}

func (w *StdoutIntercept) Write(p []byte) (int, error) {
  //w.b = p
  return os.Stdout.Write(p)
}

func (h *Handler) Eval(line string) (string, string) {
  // upon eval the Stdin should be unblocked
  if strings.TrimSpace(line) != "" {
    endCmd := make(chan []byte)

    //w := &StdoutIntercept{}

    go func() {
      h.t.UnsetRaw()

      // test the hijack function
      fields := strings.Fields(line)
      cmd := exec.Command(fields[0], fields[1:]...)
      /*stdin, err := cmd.StdinPipe()
      if err != nil {
        return err.Error(), ""
      }

      stdout, err := cmd.StdoutPipe()
      if err != nil {
        return err.Error(), ""
      }*/


      // XXX: we can't intercept the Stdout bytes in any way (because cmd depends on the Stdout file descriptor
      cmd.Stdout = os.Stdout
      //cmd.Stdout = os.Stdout
      cmd.Stdin = os.Stdin
      cmd.Stderr = os.Stderr

      /*stdoutPipe, err := cmd.StdoutPipe()
      if err != nil {
        panic(err)
      }*/

      cmd.Start()
      //if err := cmd.Start(); err != nil {
        //return err.Error(), ""
      //}

      cmd.Wait()
      //if err := cmd.Wait(); err != nil {
        //return err.Error(), ""
      //}

      h.t.SetRaw()

      //outMsg, _ := ioutil.ReadAll(stdoutPipe)

      endCmd <- []byte{}
    }()

    <- endCmd
    //return string(w.b), line

    return "", line
  } else {
    return line, line
  }
}

func ownTermLoop(cmdArgs CmdArgs) error {
  h := &Handler{}

  t := terminal.NewTerminal(h)

  h.t = t

  return t.Run()
}

// XXX: should the repl be able to handle tags?
// its main use is exploring json content
// although this can just be done with nodejs
// but that is slow to start
func main() {
  cmdArgs := parseArgs()

  if !cmdArgs.quiet {
    fmt.Println("wt-repl v0.5.1")
    fmt.Println("Type \"?\" for more information")
  }

  //if err := termLoop(cmdArgs); err != nil {
    //printMessageAndExit(err.Error())
  //}

  if err := ownTermLoop(cmdArgs); err != nil {
    printMessageAndExit(err.Error())
  }
}
