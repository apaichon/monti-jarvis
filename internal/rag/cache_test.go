package rag

import "testing"

func TestPreloadCacheRoundTrip(t *testing.T) {
	c := newPreloadCache()
	want := Result{ContextBlock: "kb"}
	c.set("demo:max:billing", want)
	got, ok := c.get("demo:max:billing")
	if !ok || got.ContextBlock != want.ContextBlock {
		t.Fatalf("cache get = %+v, ok=%v", got, ok)
	}
}