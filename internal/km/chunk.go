package km

import (
	"strings"
	"unicode/utf8"
)

const defaultChunkSize = 900

type Chunk struct {
	Index   int
	Content string
}

func ChunkText(text string, maxRunes int) []Chunk {
	if maxRunes <= 0 {
		maxRunes = defaultChunkSize
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	paragraphs := splitParagraphs(text)
	var chunks []Chunk
	var buf strings.Builder
	idx := 0

	flush := func() {
		content := strings.TrimSpace(buf.String())
		if content == "" {
			return
		}
		chunks = append(chunks, Chunk{Index: idx, Content: content})
		idx++
		buf.Reset()
	}

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		if utf8.RuneCountInString(para) > maxRunes {
			flush()
			for _, part := range splitLongParagraph(para, maxRunes) {
				chunks = append(chunks, Chunk{Index: idx, Content: part})
				idx++
			}
			continue
		}
		if buf.Len() > 0 && utf8.RuneCountInString(buf.String())+utf8.RuneCountInString(para) > maxRunes {
			flush()
		}
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(para)
	}
	flush()
	return chunks
}

func splitParagraphs(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.Split(text, "\n\n")
}

func splitLongParagraph(text string, maxRunes int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var out []string
	var buf strings.Builder
	for _, word := range words {
		next := word
		if buf.Len() > 0 {
			next = " " + word
		}
		if utf8.RuneCountInString(buf.String())+utf8.RuneCountInString(next) > maxRunes {
			out = append(out, strings.TrimSpace(buf.String()))
			buf.Reset()
			buf.WriteString(word)
			continue
		}
		buf.WriteString(next)
	}
	if buf.Len() > 0 {
		out = append(out, strings.TrimSpace(buf.String()))
	}
	return out
}