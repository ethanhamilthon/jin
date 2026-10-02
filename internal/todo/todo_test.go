package todo

import (
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	items := []Item{{"a", Pending}, {"b", InProgress}, {"c", Done}}
	got, err := Parse(Format(items))
	if err != nil || !Equal(got, items) {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestParse(t *testing.T) {
	got, err := Parse("<!-- c -->\n\n* [X] one\n  - [~]   two  words\r\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []Item{{"one", Done}, {"two words", InProgress}}
	if !Equal(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestParseEmpty(t *testing.T) {
	got, err := Parse("")
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestParseErrorLine(t *testing.T) {
	_, err := Parse("- [ ] ok\n\n- broken\n")
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidate(t *testing.T) {
	got, err := Validate([]Item{{"a\nb", ""}})
	if err != nil || got[0].Text != "a b" || got[0].Status != Pending {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := Validate([]Item{{" ", Pending}}); err == nil {
		t.Fatal("empty text accepted")
	}
	if _, err := Validate([]Item{{"a", "nope"}}); err == nil {
		t.Fatal("bad status accepted")
	}
}

func TestAllDone(t *testing.T) {
	if AllDone(nil) || AllDone([]Item{{"a", Pending}}) || !AllDone([]Item{{"a", Done}}) {
		t.Fatal("AllDone wrong")
	}
}
