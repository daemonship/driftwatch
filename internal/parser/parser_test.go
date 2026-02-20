package parser_test

import (
	"testing"

	"github.com/daemonship/driftwatch/internal/parser"
)

// Minimal valid terraform plan JSON with one updated resource.
const planWithOneUpdate = `{
  "format_version": "1.2",
  "resource_changes": [
    {
      "address": "aws_instance.web",
      "type": "aws_instance",
      "name": "web",
      "change": {
        "actions": ["update"],
        "before": {
          "ami": "ami-0abc123",
          "instance_type": "t2.micro"
        },
        "after": {
          "ami": "ami-0def456",
          "instance_type": "t3.micro"
        },
        "after_unknown": {}
      }
    }
  ]
}`

// Plan with no changes (exit code 0 scenario).
const planNoChanges = `{
  "format_version": "1.2",
  "resource_changes": []
}`

// Plan with a resource being deleted.
const planWithDelete = `{
  "format_version": "1.2",
  "resource_changes": [
    {
      "address": "aws_s3_bucket.old",
      "type": "aws_s3_bucket",
      "name": "old",
      "change": {
        "actions": ["delete"],
        "before": {
          "bucket": "my-old-bucket",
          "region": "us-east-1"
        },
        "after": null,
        "after_unknown": {}
      }
    }
  ]
}`

// Plan with create+delete (replace).
const planWithReplace = `{
  "format_version": "1.2",
  "resource_changes": [
    {
      "address": "aws_instance.db",
      "type": "aws_instance",
      "name": "db",
      "change": {
        "actions": ["delete", "create"],
        "before": {
          "ami": "ami-old",
          "instance_type": "t2.large"
        },
        "after": {
          "ami": "ami-new",
          "instance_type": "t3.large"
        },
        "after_unknown": {}
      }
    }
  ]
}`

// Plan with no-op resource (should be excluded from results).
const planWithNoOp = `{
  "format_version": "1.2",
  "resource_changes": [
    {
      "address": "aws_instance.unchanged",
      "type": "aws_instance",
      "name": "unchanged",
      "change": {
        "actions": ["no-op"],
        "before": {"ami": "ami-same"},
        "after": {"ami": "ami-same"},
        "after_unknown": {}
      }
    }
  ]
}`

func TestParse_OneUpdate(t *testing.T) {
	plan, err := parser.Parse([]byte(planWithOneUpdate))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) != 1 {
		t.Errorf("ResourceChanges count = %d, want 1", len(plan.ResourceChanges))
	}
	rc := plan.ResourceChanges[0]
	if rc.Address != "aws_instance.web" {
		t.Errorf("ResourceChanges[0].Address = %q, want %q", rc.Address, "aws_instance.web")
	}
	if rc.Action != parser.ActionUpdate {
		t.Errorf("ResourceChanges[0].Action = %q, want %q", rc.Action, parser.ActionUpdate)
	}
}

func TestParse_AttributeChanges(t *testing.T) {
	plan, err := parser.Parse([]byte(planWithOneUpdate))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) == 0 {
		t.Fatal("Parse() returned no resource changes")
	}
	rc := plan.ResourceChanges[0]
	amiChange, ok := rc.AttributeChanges["ami"]
	if !ok {
		t.Fatal("AttributeChanges missing 'ami' key")
	}
	if amiChange.Before != "ami-0abc123" {
		t.Errorf("ami Before = %v, want %q", amiChange.Before, "ami-0abc123")
	}
	if amiChange.After != "ami-0def456" {
		t.Errorf("ami After = %v, want %q", amiChange.After, "ami-0def456")
	}
}

func TestParse_NoChanges(t *testing.T) {
	plan, err := parser.Parse([]byte(planNoChanges))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) != 0 {
		t.Errorf("ResourceChanges count = %d, want 0 for no-change plan", len(plan.ResourceChanges))
	}
}

func TestParse_Delete(t *testing.T) {
	plan, err := parser.Parse([]byte(planWithDelete))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) != 1 {
		t.Errorf("ResourceChanges count = %d, want 1", len(plan.ResourceChanges))
	}
	rc := plan.ResourceChanges[0]
	if rc.Action != parser.ActionDelete {
		t.Errorf("Action = %q, want %q", rc.Action, parser.ActionDelete)
	}
	// After should be nil/zero for deletes.
	bucketChange, ok := rc.AttributeChanges["bucket"]
	if !ok {
		t.Fatal("AttributeChanges missing 'bucket' key")
	}
	if bucketChange.Before != "my-old-bucket" {
		t.Errorf("bucket Before = %v, want %q", bucketChange.Before, "my-old-bucket")
	}
}

func TestParse_Replace(t *testing.T) {
	plan, err := parser.Parse([]byte(planWithReplace))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) != 1 {
		t.Errorf("ResourceChanges count = %d, want 1", len(plan.ResourceChanges))
	}
	rc := plan.ResourceChanges[0]
	if rc.Action != parser.ActionReplace {
		t.Errorf("Action = %q, want %q", rc.Action, parser.ActionReplace)
	}
}

func TestParse_NoOpExcluded(t *testing.T) {
	plan, err := parser.Parse([]byte(planWithNoOp))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) != 0 {
		t.Errorf("ResourceChanges count = %d, want 0 (no-op resources excluded)", len(plan.ResourceChanges))
	}
}

func TestParse_FormatVersion(t *testing.T) {
	plan, err := parser.Parse([]byte(planNoChanges))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if plan.FormatVersion != "1.2" {
		t.Errorf("FormatVersion = %q, want %q", plan.FormatVersion, "1.2")
	}
}

func TestParse_MalformedJSON(t *testing.T) {
	_, err := parser.Parse([]byte(`{not valid json`))
	if err == nil {
		t.Error("Parse() error = nil, want error for malformed JSON")
	}
}

func TestParse_EmptyInput(t *testing.T) {
	_, err := parser.Parse([]byte(``))
	if err == nil {
		t.Error("Parse() error = nil, want error for empty input")
	}
}

func TestParse_MissingFormatVersion(t *testing.T) {
	// Plan JSON without format_version should still parse (field may be optional).
	input := `{"resource_changes":[]}`
	plan, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil for plan without format_version", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil for plan without format_version")
	}
}

func TestParse_OnlyDifferentAttributesReturned(t *testing.T) {
	// When before and after have the same value for a field, it should not appear
	// in AttributeChanges (only changed attributes).
	plan, err := parser.Parse([]byte(planWithOneUpdate))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if plan == nil {
		t.Fatal("Parse() returned nil plan")
	}
	if len(plan.ResourceChanges) == 0 {
		t.Fatal("Parse() returned no resource changes")
	}
	rc := plan.ResourceChanges[0]
	// Both ami and instance_type changed in the fixture.
	if len(rc.AttributeChanges) != 2 {
		t.Errorf("AttributeChanges count = %d, want 2 (only changed attributes)", len(rc.AttributeChanges))
	}
}
