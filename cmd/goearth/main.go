package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	errSilent = errors.New("")
	cmdName   string
)

func init() {
	cmdName = os.Args[0]
	if filepath.IsAbs(cmdName) {
		cmdName = filepath.Base(cmdName)
	}
	cmdName = strings.TrimSuffix(cmdName, ".exe")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n  Available commands:\n    new\n", cmdName)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
		return
	}
	cmd, args := args[0], args[1:]

	var err error

	switch cmd {
	case "new":
		err = runNew(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", cmd)
		err = flag.ErrHelp
	}

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flag.Usage()
		} else if err != errSilent {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(1)
	}
}

func runNew(args []string) (err error) {
	var opts struct {
		mod     string
		dir     string
		client  string
		title   string
		desc    string
		author  string
		version string
	}

	flags := flag.NewFlagSet(fmt.Sprintf("%s new", cmdName), flag.ExitOnError)
	flags.StringVar(&opts.dir, "d", "", "The output directory for the extension.")
	flags.StringVar(&opts.mod, "m", "", "The module name.")
	flags.StringVar(&opts.title, "title", "New extension", "The title of the extension.")
	flags.StringVar(&opts.desc, "desc", "", "The description of the extension.")
	flags.StringVar(&opts.author, "author", "", "The author of the extension.")
	flags.StringVar(&opts.version, "version", "1.0", "The author of the extension.")
	flags.StringVar(&opts.client, "c", "flash", "Target client for the extension. [flash, shockwave]")
	flags.Parse(args)

	if opts.dir == "" {
		if opts.mod != "" {
			opts.dir = path.Base(opts.mod)
		} else if opts.title != "" {
			opts.dir = opts.title
		} else {
			fmt.Fprintln(os.Stderr, "No output directory provided.")
			flags.Usage()
			return errSilent
		}
	}

	if opts.mod == "" {
		opts.mod = strings.ToLower(filepath.Base(opts.dir))
		opts.mod = regexp.MustCompile(`\s+`).ReplaceAllString(opts.mod, "")
	}

	var msgPackage, inChatIdentifiers, outWaveArgs string
	switch opts.client {
	case "flash":
		msgPackage = "xabbo.b7c.io/goearth"
		inChatIdentifiers = "in.Chat, in.Whisper, in.Shout"
		outWaveArgs = "out.AvatarExpression, 1"
	case "shockwave":
		msgPackage = "xabbo.b7c.io/goearth/shockwave"
		inChatIdentifiers = "in.CHAT, in.CHAT_2, in.CHAT_3"
		outWaveArgs = "out.WAVE"
	default:
		err = fmt.Errorf("unknown client: %q", opts.client)
		return
	}

	_, err = os.Stat(opts.dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return
		}
		err = nil
	} else {
		err = fmt.Errorf("output directory %q already exists", opts.dir)
		return
	}

	if err := os.MkdirAll(opts.dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	defer func() {
		if err != nil {
			// clean up on fail
			os.RemoveAll(opts.dir)
		}
	}()

	if opts.author == "" {
		user, err := user.Current()
		if err == nil && user.Name != "" {
			fmt.Printf("No author provided, using %q\n", user.Name)
			opts.author = user.Name
		}
	}

	data := TemplateBasicData{
		Module:            opts.mod,
		MsgPackage:        msgPackage,
		Title:             opts.title,
		Description:       opts.desc,
		Author:            opts.author,
		Version:           opts.version,
		InChatIdentifiers: inChatIdentifiers,
		OutWaveArgs:       outWaveArgs,
	}

	tmpl, err := template.New("basic").Parse(templateBasic)
	if err != nil {
		return
	}

	tmplBuf := &bytes.Buffer{}
	err = tmpl.Execute(tmplBuf, data)
	if err != nil {
		return
	}

	mainPath := filepath.Join(opts.dir, "main.go")
	err = os.WriteFile(mainPath, tmplBuf.Bytes(), 0755)
	if err != nil {
		return
	}

	err = runCommand(opts.dir, "go", "mod", "init", opts.mod)
	if err != nil {
		return
	}

	err = runCommand(opts.dir, "go", "mod", "tidy", "-v")
	if err != nil {
		return
	}

	fmt.Printf("Extension created in %q\n", opts.dir)
	return
}

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
