package schema

import (
	"strings"
	"testing"

	"github.com/yunxiao-cli/yunxiao/internal/risk"
)

func TestFindAndList(t *testing.T) {
	m := Find("codeup.mrs.create")
	if m == nil || m.Risk != risk.HighRiskWrite {
		t.Fatalf("%+v", m)
	}
	for _, id := range []string{"codeup.mrs.merge", "codeup.mrs.close", "codeup.mrs.reopen", "codeup.files.delete", "pipeline.run.cancel", "pipeline.job.retry", "pipeline.job.pass", "pipeline.job.refuse", "pipeline.create", "pipeline.update", "workitem.delete", "packages.artifacts.delete", "appstack.tags.create", "appstack.tags.bind", "appstack.variable_groups.create", "codeup.branches.create", "codeup.branches.delete", "appstack.apps.create", "versions.delete", "codeup.repos.create", "codeup.tags.create", "codeup.tags.delete", "codeup.protected_branches.create", "codeup.protected_branches.delete", "pipeline.vm_deploy.stop", "pipeline.resource_members.create", "appstack.deploy.add_hosts"} {
		mm := Find(id)
		if mm == nil || mm.Risk != risk.HighRiskWrite {
			t.Fatalf("%s %+v", id, mm)
		}
	}
	if Find("workitem.create") == nil || Find("workitem.create").Risk != risk.Write {
		t.Fatal("workitem.create")
	}
	if Find("workitem.transition") == nil || Find("workitem.transition").Risk != risk.Write {
		t.Fatal("workitem.transition")
	}
	if Find("codeup.tags.list") == nil || Find("codeup.protected_branches.list") == nil {
		t.Fatal("codeup tags/protect reads")
	}
	for _, id := range []string{"codeup.mrs.comments.create", "codeup.mrs.labels.attach", "testhub.results.update", "workitem.relations.create", "workitem.relations.delete"} {
		mm := Find(id)
		if mm == nil || mm.Risk != risk.Write {
			t.Fatalf("%s %+v", id, mm)
		}
	}
	if Find("appstack.orchestrations.list") == nil || Find("appstack.change_orders.job_logs") == nil {
		t.Fatal("appstack reads")
	}
	if Find("pipeline.get") == nil || Find("workitem.attachments.create") == nil || Find("workitem.attachments.create").Risk != risk.Write {
		t.Fatal("v0.7 reads/writes")
	}
	if Find("sprint.list") == nil || Find("organization.departments.list") == nil || Find("testhub.cases.search") == nil {
		t.Fatal("v0.8 gaps")
	}
	if Find("programs.search") == nil || Find("pipeline.vm_deploy.get") == nil || Find("appstack.release_workflows.list") == nil || Find("codeup.repos.create") == nil || Find("codeup.repos.create").Risk != risk.HighRiskWrite {
		t.Fatal("v0.9 gaps")
	}
	list := List("pipeline")
	if len(list) < 3 {
		t.Fatalf("pipeline methods=%d", len(list))
	}
	if Find("nope") != nil {
		t.Fatal()
	}
}

func TestFindWorkitemSearchAliases(t *testing.T) {
	m := Find("workitem.search")
	if m == nil || m.ID != "workitem.search" {
		t.Fatalf("%+v", m)
	}
	for _, alias := range []string{"project.searchWorkitems", "search_workitems", "searchWorkitems"} {
		a := Find(alias)
		if a == nil || a.ID != "workitem.search" {
			t.Fatalf("alias %s -> %+v", alias, a)
		}
	}
	hasAsItems := false
	for _, param := range m.Params {
		if param.Name == "as-items" {
			hasAsItems = true
		}
	}
	if !hasAsItems {
		t.Fatal("workitem.search missing as-items param")
	}
}

func TestFindUpdate(t *testing.T) {
	m := Find("update")
	if m == nil || m.ID != "update" || m.Risk != risk.Write || m.Domain != "cli" {
		t.Fatalf("%+v", m)
	}
	a := Find("self-update")
	if a == nil || a.ID != "update" {
		t.Fatalf("self-update alias -> %+v", a)
	}
	hasCheck := false
	for _, param := range m.Params {
		if param.Name == "check" {
			hasCheck = true
		}
	}
	if !hasCheck {
		t.Fatal("update missing check param")
	}
}

func TestMrsCreateReviewerParam(t *testing.T) {
	m := Find("codeup.mrs.create")
	if m == nil {
		t.Fatal("missing codeup.mrs.create")
	}
	hasReviewer := false
	for _, param := range m.Params {
		if param.Name == "reviewer" {
			hasReviewer = true
			if param.Required {
				t.Fatal("reviewer should be optional")
			}
		}
	}
	if !hasReviewer {
		t.Fatal("codeup.mrs.create missing reviewer param")
	}
	if m.Example == "" || !strings.Contains(m.Example, "--reviewer") {
		t.Fatalf("example should mention --reviewer: %q", m.Example)
	}
}
