package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"
)

//go:embed template.md
var tString string

type USN struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func main() {
	var config struct {
		BuildImage   string
		RunImage     string
		BuildDiff    string
		RunDiff      string
		PatchedJSON  string
		PatchedArray []USN
	}

	flag.StringVar(&config.BuildImage, "build-image", "", "Registry location of stack build image")
	flag.StringVar(&config.RunImage, "run-image", "", "Registry location of stack run image")
	flag.StringVar(&config.BuildDiff, "build-diff", "", "Diff of build image package receipt")
	flag.StringVar(&config.RunDiff, "run-diff", "", "Diff of run image package receipt")
	flag.StringVar(&config.PatchedJSON, "patched-usns", "", "JSON Array of patched USNs")
	flag.Parse()

	if config.PatchedJSON == "" {
		config.PatchedJSON = `[]`
	}

	err := json.Unmarshal([]byte(config.PatchedJSON), &config.PatchedArray)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("template.md").Parse(tString)
	if err != nil {
		log.Fatal(err)
	}

	b := bytes.NewBuffer(nil)
	err = t.Execute(b, config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("::set-output name=release_body::%s", escape(b.String())))
}

func escape(original string) string {
	newline := regexp.MustCompile(`\n`)
	cReturn := regexp.MustCompile(`\r`)

	result := strings.ReplaceAll(original, `%`, `%25`)
	result = newline.ReplaceAllString(result, `%0A`)
	result = cReturn.ReplaceAllString(result, `%0D`)

	return result
}
