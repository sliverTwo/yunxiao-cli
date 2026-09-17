package profile

import (
	"strings"
	"testing"
)

func TestResolveWorkflowFromWorkflows(t *testing.T) {
	p := &Profile{
		Name: "play",
		Workflows: map[string]WorkitemWorkflow{
			"task1": {
				Category: "Task",
				Name:     "任务",
				Statuses: map[string]string{"待处理": "100005", "处理中": "100010"},
				Edges:    map[string][]string{"100005": {"100010"}},
			},
		},
	}
	wf, err := p.ResolveWorkflow("task1")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Source != "workflows" || wf.Category != "Task" {
		t.Fatalf("%+v", wf)
	}
	if len(wf.Edges["100005"]) != 1 {
		t.Fatalf("edges %+v", wf.Edges)
	}
	all := wf.AllStatusIDs()
	if !all["100005"] || !all["100010"] {
		t.Fatalf("all=%v", all)
	}
}

func TestResolveWorkflowBugFallback(t *testing.T) {
	p := &Profile{
		Name:      "play",
		BugTypeID: "bug1",
		BugStatuses: map[string]string{
			"confirm":    "28",
			"processing": "100010",
		},
		BugEdges: map[string][]string{"28": {"100010"}},
	}
	wf, err := p.ResolveWorkflow("bug1")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Source != "bug_legacy" || wf.Category != "Bug" {
		t.Fatalf("%+v", wf)
	}
}

func TestResolveWorkflowMissing(t *testing.T) {
	p := &Profile{Name: "play", Workflows: map[string]WorkitemWorkflow{}}
	_, err := p.ResolveWorkflow("unknown")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "workitem update") || !strings.Contains(err.Error(), "--status") {
		t.Fatalf("expected one-off update hint: %v", err)
	}
}
