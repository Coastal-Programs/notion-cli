package commands

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestParseAPIInlineArgs_BodyQueryAndHeader(t *testing.T) {
	parsed, err := parseAPIInlineArgs([]string{
		"query=roadmap",
		"page_size:=10",
		"start_cursor==abc",
		"Accept:application/json",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := parsed.Query.Get("start_cursor"); got != "abc" {
		t.Fatalf("query start_cursor = %q", got)
	}
	if got := parsed.Headers.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept header = %q", got)
	}
	body, err := buildAPIBody(parsed.BodyFields)
	if err != nil {
		t.Fatalf("build body failed: %v", err)
	}
	if body["query"] != "roadmap" {
		t.Fatalf("query body = %#v", body["query"])
	}
	if body["page_size"] != float64(10) {
		t.Fatalf("page_size body = %#v", body["page_size"])
	}
}

func TestBuildAPIBody_NestedBracketDotAndAppend(t *testing.T) {
	parsed, err := parseAPIInlineArgs([]string{
		"parent[page_id]=abc123",
		"properties.Name.title[0].text.content=Meeting notes",
		"rich_text[][text][content]=First",
		"rich_text[][text][content]=Second",
		"filter:={\"property\":\"object\",\"value\":\"page\"}",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	body, err := buildAPIBody(parsed.BodyFields)
	if err != nil {
		t.Fatalf("build body failed: %v", err)
	}
	b, _ := json.Marshal(body)
	var got map[string]any
	_ = json.Unmarshal(b, &got)

	parent := got["parent"].(map[string]any)
	if parent["page_id"] != "abc123" {
		t.Fatalf("parent = %#v", parent)
	}
	props := got["properties"].(map[string]any)
	name := props["Name"].(map[string]any)
	title := name["title"].([]any)
	text := title[0].(map[string]any)["text"].(map[string]any)
	if text["content"] != "Meeting notes" {
		t.Fatalf("nested title content = %#v", text["content"])
	}
	richText := got["rich_text"].([]any)
	if len(richText) != 2 {
		t.Fatalf("rich_text len = %d", len(richText))
	}
	filter := got["filter"].(map[string]any)
	if filter["value"] != "page" {
		t.Fatalf("typed filter = %#v", filter)
	}
}

func TestAPIFieldsAsFormFields_JSONValues(t *testing.T) {
	fields, err := apiFieldsAsFormFields([]apiInlineField{
		{Path: "part_number", Value: float64(1)},
		{Path: "name", Value: "chunk"},
	})
	if err != nil {
		t.Fatalf("form fields failed: %v", err)
	}
	want := map[string]string{"part_number": "1", "name": "chunk"}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %#v, want %#v", fields, want)
	}
}

func TestLooksLikeHeaderArg(t *testing.T) {
	if !looksLikeHeaderArg("X-Trace-Id:abc") {
		t.Fatal("expected header arg")
	}
	if looksLikeHeaderArg("filter:={\"a\":1}") {
		t.Fatal("typed body arg should not be a header")
	}
	if looksLikeHeaderArg("query==roadmap") {
		t.Fatal("query arg should not be a header")
	}
}

func TestParseAPIInlineArgs_RepeatedHeaders(t *testing.T) {
	parsed, err := parseAPIInlineArgs([]string{"X-Test:a", "X-Test:b"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := parsed.Headers.Values("X-Test"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("headers = %#v", got)
	}
	if parsed.Headers.Get("X-Test") == "" || parsed.Headers.Get(http.CanonicalHeaderKey("X-Test")) == "" {
		t.Fatalf("header canonicalization failed: %#v", parsed.Headers)
	}
}
