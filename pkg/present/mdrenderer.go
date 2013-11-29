package present

import (
	"bytes"
	"github.com/russross/blackfriday"
	"html/template"
	"log"
	"strings"
)

type PresentContent struct {
	htmlRender        blackfriday.Renderer
	sections          []*Section
	parentSection     *Section
	lastSection       *Section
	lastList          []string
	lastSectionNumber []int
	textBuffer        *Text
}

func PresentContentRenderer(flags int) (r *PresentContent) {
	rd := blackfriday.HtmlRenderer(flags, "", "")
	return &PresentContent{htmlRender: rd}
}

func (pc *PresentContent) Sections() (r []Section) {
	for _, s := range pc.sections {
		r = append(r, *s)
	}
	return
}

func (pc *PresentContent) Header(out *bytes.Buffer, text func() bool, level int) {
	// pc.htmlRender.Header(content, text, level)
	content, extracted := extractText(out, text)
	if !extracted {
		return
	}

	if pc.lastSection != nil && pc.textBuffer != nil {
		pc.lastSection.Elem = append(pc.lastSection.Elem, *pc.textBuffer)
		pc.textBuffer = nil
	}

	pc.lastSection = &Section{
		Number: levelNumber(pc.lastSectionNumber, level),
		Title:  content,
	}
	pc.lastSectionNumber = pc.lastSection.Number
	pc.sections = append(pc.sections, pc.lastSection)

	if pc.parentSection == nil || len(pc.parentSection.Number) < level {
		pc.parentSection = pc.lastSection
	}
	log.Println("Header", content)
}

func extractText(out *bytes.Buffer, text func() bool) (r string, extracted bool) {
	marker := out.Len()
	if !text() {
		out.Truncate(marker)
		return
	}
	extracted = true
	r = string(out.Bytes()[marker:])
	return
}

func levelNumber(lastSectionNumber []int, level int) (r []int) {
	if len(lastSectionNumber) >= level {
		r = append(r, lastSectionNumber[0:level-1]...)
		last := lastSectionNumber[level-1]
		r = append(r, last+1)
		return
	}

	for i := 0; i < level; i++ {
		if len(lastSectionNumber) > i {
			r = append(r, lastSectionNumber[i])
		} else {
			r = append(r, 1)
		}
	}
	return
}

func (pc *PresentContent) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	code := Code{
		Text: template.HTML(cleanEscapeHighlightCode(text, "")),
	}

	pc.lastSection.Elem = append(pc.lastSection.Elem, code)
	log.Println("BlockCode", string(text), lang)
	return
}

func (pc *PresentContent) List(out *bytes.Buffer, text func() bool, flags int) {
	_, extracted := extractText(out, text)
	if !extracted {
		return
	}
	list := List{
		Bullet: pc.lastList,
	}
	pc.lastSection.Elem = append(pc.lastSection.Elem, list)
	pc.lastList = []string{}
	log.Println("List", flags)
	return
}

func (pc *PresentContent) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if pc.textBuffer != nil {
		if len(pc.textBuffer.Lines) == 0 || pc.textBuffer.Lines[0] != string(text) {
			pc.lastSection.Elem = append(pc.lastSection.Elem, *pc.textBuffer)
		}
		pc.textBuffer = nil
	}

	pc.lastList = append(pc.lastList, string(text))
	log.Println("ListItem", string(text), flags)
	return
}

func (pc *PresentContent) Paragraph(out *bytes.Buffer, text func() bool) {
	content, extracted := extractText(out, text)
	if !extracted {
		return
	}

	if pc.textBuffer != nil {
		pc.lastSection.Elem = append(pc.lastSection.Elem, *pc.textBuffer)
		pc.textBuffer = nil
	}

	txt := Text{
		Lines: splitLines(content),
	}
	pc.textBuffer = &txt
	// pc.lastSection.Elem = append(pc.lastSection.Elem, txt)
	log.Println("Paragraph", string(out.Bytes()))
	return
}

func splitLines(txt string) (lines []string) {
	txt = strings.Replace(txt, "\r", "", -1)
	lines = strings.Split(txt, "\n")
	return
}

func (pc *PresentContent) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	pc.Link(out, link, nil, nil)
	return
}

func (pc *PresentContent) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("`")
	out.WriteString(strings.Replace(string(text), " ", "`", -1))
	out.WriteString("`")
	log.Println("CodeSpan", string(text))
	return
}

func (pc *PresentContent) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	pc.Emphasis(out, text)
	log.Println("DoubleEmphasis", string(text))
	return
}

func (pc *PresentContent) Emphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("*")
	out.WriteString(strings.Replace(string(text), " ", "*", -1))
	out.WriteString("*")
	log.Println("Emphasis", string(text))
	return
}

func (pc *PresentContent) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	img := Image{
		URL: string(link),
	}
	pc.lastSection.Elem = append(pc.lastSection.Elem, img)
	log.Println("Image", string(link), string(title), string(alt))
	return
}

func (pc *PresentContent) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	// pc.htmlRender.Link(out, link, title, content)
	//[[url][label]]
	out.WriteString("[[")
	out.Write(link)
	out.WriteString("][")
	out.Write(content)
	out.WriteString("]]")
	log.Println("Link", string(link), string(title), string(content))
	return
}

func (pc *PresentContent) RawHtmlTag(out *bytes.Buffer, text []byte) {
	out.Write(text)
	log.Println("RawHtmlTag", string(text))
	return
}

func (pc *PresentContent) TripleEmphasis(out *bytes.Buffer, text []byte) {
	pc.Emphasis(out, text)
	log.Println("TripleEmphasis", string(text))
	return
}

func (pc *PresentContent) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.Write(text)
	log.Println("StrikeThrough", string(text))
	return
}

func (pc *PresentContent) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
	log.Println("Entity", string(entity))
	return
}

func (pc *PresentContent) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
	log.Println("NormalText", string(text))
	return
}

// ================================================

func (pc *PresentContent) BlockHtml(out *bytes.Buffer, text []byte) {
	log.Println("BlockHtml", string(text))
	return
}

func (pc *PresentContent) HRule(out *bytes.Buffer) {
	log.Println("HRule")
	return
}

func (pc *PresentContent) LineBreak(out *bytes.Buffer) {
	return
}

func (pc *PresentContent) BlockCodeNormal(out *bytes.Buffer, text []byte, lang string) {
	log.Println("BlockCodeNormal", string(text), lang)
	return
}

func (pc *PresentContent) BlockCodeGithub(out *bytes.Buffer, text []byte, lang string) {
	log.Println("BlockCodeGithub", string(text), lang)
	return
}

func (pc *PresentContent) BlockQuote(out *bytes.Buffer, text []byte) {
	log.Println("BlockQuote", string(text))
	return
}

func (pc *PresentContent) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	log.Println("Table", string(header), string(body), columnData)
	return
}

func (pc *PresentContent) TableRow(out *bytes.Buffer, text []byte) {
	log.Println("TableRow", string(text))
	return
}

func (pc *PresentContent) TableCell(out *bytes.Buffer, text []byte, align int) {
	log.Println("TableCell", string(text), align)
	return
}

func (pc *PresentContent) Smartypants(out *bytes.Buffer, text []byte) {
	log.Println("Smartypants")
	return
}

func (pc *PresentContent) DocumentHeader(out *bytes.Buffer) {
	log.Println("DocumentHeader")
	return
}

func (pc *PresentContent) DocumentFooter(out *bytes.Buffer) {
	log.Println("DocumentFooter")
	return
}

func (pc *PresentContent) TocHeader(text []byte, level int) {
	log.Println("TocHeader", string(text), level)
	return
}

func (pc *PresentContent) TocFinalize() {
	log.Println("TocFinalize")
	return
}

func (pc *PresentContent) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	log.Println("FootnoteItem")
	return
}

func (pc *PresentContent) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	log.Println("FootnoteRef")
	return
}

func (pc *PresentContent) Footnotes(out *bytes.Buffer, text func() bool) {
	log.Println("Footnotes")
	return
}
