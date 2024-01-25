package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/exp/slices"
)

type Dir string

const (
	Out Dir = "Out"
	In  Dir = "In"
)

type Release struct {
	Id      int     `json:"id"`
	Variant int     `json:"variant"`
	Version string  `json:"version"`
	Size    int     `json:"size"`
	Packets Packets `json:"packets"`
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

func main() {
	log.SetPrefix("messages: ")
	err := run()
	if err != nil {
		log.Printf("Error: %s", err)
	}
}

func run() (err error) {
	log.Printf("Generating message variables")

	dir, err := os.Getwd()
	if err != nil {
		return
	}
	log.Printf("Current directory: %s", dir)

	log.Printf("Fetching latest client release...")
	res, err := http.Get("https://api.sulek.dev/releases?variant=flash-windows")
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("server responded %s", res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	releases := []Release{}
	err = json.Unmarshal(data, &releases)
	if err != nil {
		return
	}

	if len(releases) == 0 {
		return fmt.Errorf("no releases found")
	}

	release := releases[0]
	log.Printf("Using client version %q", release.Version)

	log.Printf("Fetching client messages...")
	res, err = http.Get("https://api.sulek.dev/releases/" + release.Version + "/messages")
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("server responded %s", res.Status)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}

	container := struct {
		Messages Messages `json:"messages"`
	}{}
	err = json.Unmarshal(data, &container)
	if err != nil {
		return
	}

	messages := container.Messages
	for i, msg := range messages.Incoming {
		messages.Incoming[i].Name = cleanMessageName(In, msg.Name)
	}
	for i, msg := range messages.Outgoing {
		messages.Outgoing[i].Name = cleanMessageName(Out, msg.Name)
	}
	slices.SortFunc(messages.Incoming, sortMessage)
	slices.SortFunc(messages.Outgoing, sortMessage)

	bufferIncoming, err := generateMessagesSrc(In, release, messages.Incoming)
	if err != nil {
		return
	}

	bufferOutgoing, err := generateMessagesSrc(Out, release, messages.Outgoing)
	if err != nil {
		return
	}

	err = os.WriteFile("in/in.go", bufferIncoming, 0755)
	if err != nil {
		return
	}
	log.Printf("Wrote %d incoming messages", len(messages.Incoming))
	err = formatSrc("in/in.go")
	if err != nil {
		return
	}

	err = os.WriteFile("out/out.go", bufferOutgoing, 0755)
	if err != nil {
		return
	}
	err = formatSrc("out/out.go")
	if err != nil {
		return
	}
	log.Printf("Wrote %d outgoing messages", len(messages.Outgoing))

	return
}

func generateMessagesSrc(dir Dir, release Release, messages []Message) (buffer []byte, err error) {
	b := bytes.NewBuffer(nil)

	fmt.Fprintf(b, "// Source generated for release %s (source: Sulek API)\n\n", release.Version)
	fmt.Fprintf(b, "package %s\n\n", strings.ToLower(string(dir)))
	fmt.Fprint(b, "import g \"github.com/b7c/goearth\"\n\n")

	fmt.Fprint(b, "func id(name string) g.Identifier {\n")
	fmt.Fprintf(b, "\treturn g.Identifier{Dir: g.%s, Name: name}\n", dir)
	fmt.Fprint(b, "}\n\n")

	fmt.Fprint(b, "var (\n")
	for _, message := range messages {
		fmt.Fprintf(b, "\t%[1]s = id(%[1]q)\n", message.Name)
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

func sortMessage(a, b Message) int {
	if a.Name < b.Name {
		return -1
	}
	if a.Name > b.Name {
		return 1
	}
	return 0
}
