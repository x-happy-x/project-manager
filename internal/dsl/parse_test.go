package dsl

import (
	"reflect"
	"testing"
)

func TestSplitColonCommands_NamedSequence(t *testing.T) {
	in := []string{":build", "-DskipTests", ":run", "-d"}
	got := SplitColonCommands(in)
	want := []Chunk{
		{Name: "build", Args: []string{"-DskipTests"}},
		{Name: "run", Args: []string{"-d"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSplitColonCommands_RawMode(t *testing.T) {
	in := []string{"docker", "compose", "ps"}
	got := SplitColonCommands(in)
	if len(got) != 1 || got[0].Name != "__RAW__" {
		t.Fatalf("expected RAW, got %#v", got)
	}
	want := []string{"docker", "compose", "ps"}
	if !reflect.DeepEqual(got[0].Args, want) {
		t.Fatalf("args mismatch: got %v, want %v", got[0].Args, want)
	}
}

func TestSplitColonCommands_Empty(t *testing.T) {
	got := SplitColonCommands(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %#v", got)
	}
}
