package km

import (
	"strings"
	"testing"
)

func TestChunkTextSplitsParagraphs(t *testing.T) {
	chunks := ChunkText("Line one.\n\nLine two.", 50)
	if len(chunks) != 1 {
		t.Fatalf("len(chunks) = %d, want 1 merged chunk", len(chunks))
	}
	long := strings.Repeat("word ", 200) + "\n\n" + strings.Repeat("more ", 200)
	chunks = ChunkText(long, 50)
	if len(chunks) < 2 {
		t.Fatalf("len(chunks) = %d, want multiple chunks for long text", len(chunks))
	}
}