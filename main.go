package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"
)

// A Mediathek is an implementation of a mediathek-rtmp extractor
// taken from https://code.google.com/p/go/source/browse/src/cmd/go/main.go
type Mediathek struct {
	// Parse runs the parser
	// The args are the arguments after the site name.
	Parse func(cmd *Mediathek, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the site name.
	UsageLine string

	// Short is the short description shown in the 'gema help' output.
	Short string

	// Long is the long message shown in the 'gema help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this mediathek.
	Flag flag.FlagSet

	// CustomFlags indicates that the Mediathek will do its own
	// flag parsing.
	CustomFlags bool
}

// Name returns the mediatheks's name: the first word in the usage line.
func (m *Mediathek) Name() string {
	name := m.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (m *Mediathek) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", m.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(m.Long))
	os.Exit(2)
}

// list of available mediatheken
// The order here is the order in which they are printed by 'go help'.
var mediatheken = []*Mediathek{
	mediaArd,
	mediaZdf,
	mediaArtePlus7,
	mediaArteVideos,
}

var exitStatus = 0
var exitMu sync.Mutex

func setExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	// log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	for _, theken := range mediatheken {
		if theken.Name() == args[0] && theken.Parse != nil {
			theken.Flag.Usage = func() { theken.Usage() }
			if theken.CustomFlags {
				args = args[1:]
			} else {
				theken.Flag.Parse(args[1:])
				args = theken.Flag.Args()
			}
			theken.Parse(theken, args)
			exit()
			return
		}
	}

	fmt.Fprintf(os.Stderr, "gema: unkown media site %q\nRun 'gema help' for usage.\n", args[0])
	setExitStatus(2)
	exit()
}

var usageTemplate = `gib erstma alles (gema) is a helper for extracting rtmp urls from media sites.

usually you run it like rtmpdump -r $(gema siteName url) -o filename.

Usage:
	gema siteName [arguments]

The sites are:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use "gema help siteName" to get more information.
`

var helpTemplate = `usage: gema {{.UsageLine}}

{{.Long | trim}}
`

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, mediatheken)
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'go help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gema help site\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'go help'
	}

	arg := args[0]

	for _, cmd := range mediatheken {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'go help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'gema help'.\n", arg)
	os.Exit(2) // failed at 'go help cmd'
}

var atexitFuncs []func()

func atexit(f func()) {
	atexitFuncs = append(atexitFuncs, f)
}

func exit() {
	for _, f := range atexitFuncs {
		f()
	}
	os.Exit(exitStatus)
}
