package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

type Dir int

const (
	In Dir = iota
	Out
)

func (dir Dir) ShortName() string {
	switch dir {
	case In:
		return "in"
	case Out:
		return "out"
	default:
		return "unknown"
	}
}

func (dir Dir) Name() string {
	switch dir {
	case In:
		return "incoming"
	case Out:
		return "outgoing"
	default:
		return "unknown"
	}
}

type Release struct {
	Id          int     `json:"id"`
	Variant     int     `json:"variant"`
	VariantPath string  `json:"variantPath"`
	Version     string  `json:"version"`
	Size        int     `json:"size"`
	Packets     Packets `json:"packets"`
}

type Packets struct {
	Count int `json:"count"`
	Total int `json:"total"`
}

type MessagesContainer struct {
	Messages Messages `json:"messages"`
}

type Messages struct {
	Incoming []Message `json:"incoming"`
	Outgoing []Message `json:"outgoing"`
}

type Message struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"asNamespace"`
	Class     string `json:"asClass"`
	Confident bool   `json:"confident"`
}

var opts struct {
	dir     string
	variant string
}

func main() {
	flag.StringVar(&opts.dir, "dir", ".", "The output directory.")
	flag.StringVar(&opts.variant, "variant", "", "The client variant.")
	flag.Parse()

	opts.variant = strings.ToLower(opts.variant)

	log.SetPrefix("messages: ")
	if err := run(); err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
}

func run() (err error) {
	log.Println("Generating message variables")

	wd, err := os.Getwd()
	if err != nil {
		return
	}
	log.Printf("Current directory: %s", wd)

	inDir := filepath.Join(opts.dir, "in")
	err = os.MkdirAll(inDir, 0755)
	if err != nil {
		return
	}

	outDir := filepath.Join(opts.dir, "out")
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return
	}

	log.Println("Fetching latest client release...")

	res, err := http.Get(fmt.Sprintf("https://api.sulek.dev/releases?variant=%s", url.QueryEscape(opts.variant)))
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("server responded %s", res.Status)
	}

	releases := []Release{}
	err = json.NewDecoder(res.Body).Decode(&releases)
	if err != nil {
		return
	}

	if len(releases) == 0 {
		return fmt.Errorf("no releases found")
	}

	release := releases[0]
	log.Printf("Using client version: %s", release.Version)

	log.Printf("Fetching client messages...")
	res, err = http.Get("https://api.sulek.dev/releases/" + opts.variant + "/" + release.Version + "/messages")
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("server responded %s", res.Status)
	}

	container := struct {
		Messages Messages `json:"messages"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&container)
	if err != nil {
		return
	}

	messages := container.Messages

	if !strings.HasPrefix(opts.variant, "shockwave") {
		for i, msg := range messages.Incoming {
			messages.Incoming[i].Name = cleanMessageName(In, msg.Name)
		}
		for i, msg := range messages.Outgoing {
			messages.Outgoing[i].Name = cleanMessageName(Out, msg.Name)
		}
	}
	slices.SortFunc(messages.Incoming, compareMessageName)
	slices.SortFunc(messages.Outgoing, compareMessageName)

	err = writeMessagesSrc(inDir, In, release, messages.Incoming)
	if err != nil {
		return
	}

	err = writeMessagesSrc(outDir, Out, release, messages.Outgoing)
	if err != nil {
		return
	}

	return
}

func writeMessagesSrc(outputDir string, direction Dir, release Release, messages []Message) (err error) {
	buf, err := generateMessagesSrc(direction, release, messages)
	if err != nil {
		return
	}

	outPath := filepath.Join(outputDir, "messages.go")
	err = os.WriteFile(outPath, buf, 0755)
	if err != nil {
		return
	}

	log.Printf("Wrote %d %s messages", len(messages), direction.Name())
	err = formatSrc(outPath)
	return
}

func generateMessagesSrc(dir Dir, release Release, messages []Message) (buffer []byte, err error) {
	b := &bytes.Buffer{}

	fmt.Fprintf(b, "// Generated for %s release %s (source: sulek.dev)\n\n", release.VariantPath, release.Version)
	fmt.Fprintf(b, "package %s\n\n", dir.ShortName())
	fmt.Fprint(b, "import g \"xabbo.b7c.io/goearth\"\n\n")

	dirShortName := dir.ShortName()
	dirShortNameTitle := strings.ToUpper(dirShortName[:1]) + dirShortName[1:]

	fmt.Fprint(b, "var (\n")
	for _, message := range messages {
		fmt.Fprintf(b, "\t%[1]s = g.%s.Id(%[1]q)\n", message.Name, dirShortNameTitle)
	}
	fmt.Fprint(b, ")\n")

	buffer = b.Bytes()
	return
}

func formatSrc(name string) (err error) {
	cmd := exec.Command("go", "fmt", name)
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}
	return
}

func cleanMessageName(dir Dir, name string) (result string) {
	var cut bool
	switch dir {
	case In:
		result, cut = strings.CutSuffix(name, "Event")
	case Out:
		result, cut = strings.CutSuffix(name, "Composer")
	}
	if cut {
		result = strings.TrimSuffix(result, "Message")
	}
	return
}

func compareMessageName(a, b Message) int {
	if a.Name < b.Name {
		return -1
	}
	if a.Name > b.Name {
		return 1
	}
	return 0
}
