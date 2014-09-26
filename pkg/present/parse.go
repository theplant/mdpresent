// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package present

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/theplant/blackfriday"
)

var (
	parsers = make(map[string]func(string, int, string) (Elem, error))
	funcs   = template.FuncMap{}
)

// Template returns an empty template with the action functions in its FuncMap.
func Template() *template.Template {
	return template.New("").Funcs(funcs)
}

// Render renders the doc to the given writer using the provided template.
func (d *Doc) Render(w io.Writer, t *template.Template) error {
	data := struct {
		*Doc
		Template *template.Template
	}{d, t}
	return t.ExecuteTemplate(w, "root", data)
}

type ParseFunc func(fileName string, lineNumber int, inputLine string) (Elem, error)

// Register binds the named action, which does not begin with a period, to the
// specified parser to be invoked when the name, with a period, appears in the
// present input text.
func Register(name string, parser ParseFunc) {
	if len(name) == 0 || name[0] == ';' {
		panic("bad name in Register: " + name)
	}
	parsers["."+name] = parser
}

// Doc represents an entire document.
type Doc struct {
	Title    string
	Subtitle string
	Time     time.Time
	Authors  []Author
	Sections []Section
}

// Author represents the person who wrote and/or is presenting the document.
type Author struct {
	Elem []Elem
}

// TextElem returns the first text elements of the author details.
// This is used to display the author' name, job title, and company
// without the contact details.
func (p *Author) TextElem() (elems []Elem) {
	for _, el := range p.Elem {
		if _, ok := el.(Text); !ok {
			break
		}
		elems = append(elems, el)
	}
	return
}

// Section represents a section of a document (such as a presentation slide)
// comprising a title and a list of elements.
type Section struct {
	Number []int
	Title  string
	Elem   []Elem
}

func (s Section) Sections() (sections []Section) {
	for _, e := range s.Elem {
		if s, ok := e.(Section); ok {
			sections = append(sections, s)
		}
	}
	return
}

// Level returns the level of the given section.
// The document title is level 1, main section 2, etc.
func (s Section) Level() int {
	return len(s.Number) + 1
}

// FormattedNumber returns a string containing the concatenation of the
// numbers identifying a Section.
func (s Section) FormattedNumber() string {
	b := &bytes.Buffer{}
	for _, n := range s.Number {
		fmt.Fprintf(b, "%v.", n)
	}
	return b.String()
}

func (s Section) TemplateName() string { return "section" }

// Elem defines the interface for a present element. That is, something that
// can provide the name of the template used to render the element.
type Elem interface {
	TemplateName() string
}

// renderElem implements the elem template function, used to render
// sub-templates.
func renderElem(t *template.Template, e Elem) (template.HTML, error) {
	var data interface{} = e
	if s, ok := e.(Section); ok {
		data = struct {
			Section
			Template *template.Template
		}{s, t}
	}
	return execTemplate(t, e.TemplateName(), data)
}

func init() {
	funcs["elem"] = renderElem
}

// execTemplate is a helper to execute a template and return the output as a
// template.HTML value.
func execTemplate(t *template.Template, name string, data interface{}) (template.HTML, error) {
	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, name, data)
	if err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}

// Text represents an optionally preformatted paragraph.
type Text struct {
	Lines      []string
	originText string
	Pre        bool
}

func (t Text) TemplateName() string { return "text" }

// List represents a bulleted list.
type List struct {
	Bullet []string
}

func (l List) TemplateName() string { return "list" }

// Lines is a helper for parsing line-based input.
type Lines struct {
	line int // 0 indexed, so has 1-indexed number of last line returned
	text []string
}

func readLines(r io.Reader) (*Lines, error) {
	contentBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Lines{0, strings.Split(string(contentBytes), "\n")}, nil
}

func (l *Lines) next() (text string, ok bool) {
	current := l.line
	l.line++
	if current >= len(l.text) {
		return "", false
	}
	text = l.text[current]
	ok = true
	return
}

func (l *Lines) back() {
	l.line--
}

func (l *Lines) nextNonEmpty() (text string, ok bool) {
	for {
		text, ok = l.next()
		if !ok {
			return
		}
		if len(text) > 0 {
			break
		}
	}
	return
}

// ParseMode represents flags for the Parse function.
type ParseMode int

const (
	// If set, parse only the title and subtitle.
	TitlesOnly ParseMode = 1
)

// Parse parses the document in the file specified by name.
func Parse(r io.Reader, name string, mode ParseMode) (*Doc, error) {
	doc := new(Doc)
	lines, err := readLines(r)
	if err != nil {
		return nil, err
	}
	err = parseHeader(doc, lines)
	if err != nil {
		return nil, err
	}
	if mode&TitlesOnly != 0 {
		return doc, nil
	}
	// Authors
	if doc.Authors, err = parseAuthors(lines); err != nil {
		return nil, err
	}
	// Sections
	if doc.Sections, err = parseMarkdownSections(name, lines, []int{}, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func lineBytes(lines *Lines) (r []byte) {
	for {
		l, ok := lines.next()
		if !ok {
			break
		}
		r = append(r, '\n')
		r = append(r, []byte(l)...)
	}
	return
}

// parseSections parses Sections from lines for the section level indicated by
// number (a nil number indicates the top level).
func parseMarkdownSections(name string, lines *Lines, number []int, doc *Doc) (r []Section, err error) {
	log.Println("starting parseMarkdownSections")
	body := lineBytes(lines)

	// set up the HTML renderer
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_XHTML
	// htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	renderer := PresentContentRenderer(htmlFlags)

	// set up the parser
	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	// extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_UNICODE_LIST_ITEM
	extensions |= blackfriday.EXTENSION_NO_LIST_ITEM_BLOCK
	// extensions |= blackfriday.EXTENSION_SPACE_HEADERS  // to make sure that only "#\n\n" with empty titles also generate <h1></h1> works.
	blackfriday.Markdown(body, renderer, extensions)
	r = renderer.Sections()

	return
}

func parseHeader(doc *Doc, lines *Lines) error {
	var ok bool
	// First non-empty line starts header.
	doc.Title, ok = lines.nextNonEmpty()
	if !ok {
		return errors.New("unexpected EOF; expected title")
	}
	for {
		text, ok := lines.next()
		if !ok {
			return errors.New("unexpected EOF")
		}
		if text == "" {
			break
		}
		if t, ok := parseTime(text); ok {
			doc.Time = t
			break
		}
		if doc.Subtitle == "" {
			doc.Subtitle = text
			continue
		}
		return fmt.Errorf("unexpected header line: %q", text)
	}
	return nil
}

func parseAuthors(lines *Lines) (authors []Author, err error) {
	// This grammar demarcates authors with blanks.

	// Skip blank lines.
	if _, ok := lines.nextNonEmpty(); !ok {
		return nil, errors.New("unexpected EOF")
	}
	lines.back()

	var a *Author
	for {
		text, ok := lines.next()
		if !ok {
			return nil, errors.New("unexpected EOF")
		}

		// If we find a section heading, we're done.
		if strings.HasPrefix(text, "#") {
			lines.back()
			break
		}

		// If we encounter a blank we're done with this author.
		if a != nil && len(text) == 0 {
			authors = append(authors, *a)
			a = nil
			continue
		}
		if a == nil {
			a = new(Author)
		}

		// Parse the line. Those that
		// - begin with @ are twitter names,
		// - contain slashes are links, or
		// - contain an @ symbol are an email address.
		// The rest is just text.
		var el Elem
		switch {
		case strings.HasPrefix(text, "@"):
			el = parseURL("http://twitter.com/" + text[1:])
			if l, ok := el.(Link); ok {
				l.Label = text
				el = l
			}
		case strings.Contains(text, ":"):
			el = parseURL(text)
		case strings.Contains(text, "@"):
			el = parseURL("mailto:" + text)
		}
		if el == nil {
			el = Text{Lines: []string{text}}
		}
		a.Elem = append(a.Elem, el)
	}
	if a != nil {
		authors = append(authors, *a)
	}
	return authors, nil
}

func parseURL(text string) Elem {
	u, err := url.Parse(text)
	if err != nil {
		log.Printf("Parse(%q): %v", text, err)
		return nil
	}
	return Link{URL: u}
}

func parseTime(text string) (t time.Time, ok bool) {
	t, err := time.Parse("15:04 2 Jan 2006", text)
	if err == nil {
		return t, true
	}
	t, err = time.Parse("2 Jan 2006", text)
	if err == nil {
		// at 11am UTC it is the same date everywhere
		t = t.Add(time.Hour * 11)
		return t, true
	}
	return time.Time{}, false
}
