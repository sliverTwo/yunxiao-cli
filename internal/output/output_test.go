package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSuccessEnvelope(t *testing.T) {
	var buf bytes.Buffer
	Stdout = &buf
	JQ = ""
	Format = "json"
	if err := Success(map[string]any{"a": 1}, map[string]any{"count": 1}); err != nil {
		t.Fatal(err)
	}
	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.OK {
		t.Fatal("ok")
	}
}

func TestJQFilter(t *testing.T) {
	var buf bytes.Buffer
	Stdout = &buf
	JQ = ".data.x"
	Format = "json"
	if err := Success(map[string]any{"x": 42}, nil); err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.(float64) != 42 {
		t.Fatalf("%v", v)
	}
	JQ = ""
}

func TestFailConfirmationShape(t *testing.T) {
	var buf bytes.Buffer
	Stderr = &buf
	JQ = ""
	err := Fail(ErrorBody{
		Type: "confirmation", Subtype: "confirmation_required",
		Message: "needs yes", Hint: "add --yes", Risk: "high-risk-write", Action: "codeup mrs create",
	}, 10)
	ee, ok := err.(ExitError)
	if !ok || ee.Code != 10 {
		t.Fatalf("%v", err)
	}
	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Error == nil || env.Error.Subtype != "confirmation_required" {
		t.Fatalf("%+v", env)
	}
}
