package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type Language struct {
	ID         uint `yaml:"language_id"`
	Name       string
	Extensions []string `yaml:"extensions"`
	Filenames  []string `yaml:"filenames"`
}

// NOTE: Due to Go's random iteration of maps, generating will always change languages.go
func main() {
	log.SetFlags(0)
	log.SetPrefix("lingo: ")

	content, err := ioutil.ReadFile("languages.yml")
	if err != nil {
		panic(err)
	}
	languages := map[string]Language{}
	err = yaml.Unmarshal([]byte(content), &languages)
	if err != nil {
		panic(err)
	}

	g := Generator{}
	g.Printf("// Code generated by \"lingo\"; DO NOT EDIT.\n")
	g.Printf("\n")
	g.Printf("package lingo\n")
	g.Printf("\n")
	g.Printf("type Language struct {\n")
	g.Printf("\tID uint\n")
	g.Printf("\tName string\n")
	g.Printf("\tExtensions []string\n")
	g.Printf("\tFilenames []string\n")
	g.Printf("}\n")
	g.Printf("\n")

	g.Printf("var (\n")
	languagesByExtension := map[string][]Language{}
	languagesByFileName := map[string][]Language{}
	g.Printf("\tLanguages = map[string]Language{\n")
	for k, v := range languages {
		v.Name = k
		g.Printf("\t\t\"%s\":", k)
		g.printLanguage(&v)
		g.Printf(",\n")

		for _, e := range v.Extensions {
			x := languagesByExtension[e]
			x = append(x, v)
			languagesByExtension[e] = x
		}

		for _, e := range v.Filenames {
			x := languagesByFileName[e]
			x = append(x, v)
			languagesByFileName[e] = x
		}
	}
	g.Printf("}\n")

	// Languages by extension
	g.Printf("\tLanguagesByExtension = map[string]string{\n")
	for ext, langs := range languagesByExtension {
		lang := primaryLanguage(langs)
		if lang == nil {
			continue
		}
		g.Printf("\t\t\"%s\": \"%s\",\n", ext, lang.Name)
	}
	g.Printf("\t}\n")

	// Languages by filename
	g.Printf("\tLanguagesByFileName = map[string]string{\n")
	for name, langs := range languagesByFileName {
		lang := primaryLanguage(langs)
		if lang == nil {
			continue
		}
		g.Printf("\t\t\"%s\": \"%s\",\n", name, lang.Name)
	}
	g.Printf("\t}\n")

	// end of `var` declaration
	g.Printf(")\n")

	// Format the output.
	src := g.format()

	err = ioutil.WriteFile("languages.go", src, 0644)
	if err != nil {
		panic(err)
	}
}

// Some extensions and filenames are overloaded. Return what we consider primary
// languages. In the general case just return the first langauge.
func primaryLanguage(languages []Language) *Language {
	if len(languages) < 1 {
		return nil
	}

	if len(languages) == 1 {
		return &languages[0]
	}

	for _, l := range languages {
		if l.Name == "Markdown" {
			return &l
		}
		if l.Name == "R" {
			return &l
		}
		if l.Name == "SQL" {
			return &l
		}
		if l.Name == "TSX" {
			return &l
		}
	}

	return &languages[0]
}

func (g *Generator) printLanguage(language *Language) {
	var extensions []string
	for _, e := range language.Extensions {
		extensions = append(extensions, fmt.Sprintf(`"%s"`, e))
	}
	exts := strings.Join(extensions, ", ")
	g.Printf("Language{ID: %d, Name:\"%s\", Extensions: []string{%s} }", language.ID, language.Name, exts)
}

type Generator struct {
	buf bytes.Buffer // Accumulated output.
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}
