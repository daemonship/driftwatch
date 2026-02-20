// Package parser parses the JSON output from terraform plan -json.
// Requires Terraform >= 1.0.0.
package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Action represents the type of change to a resource.
type Action string

const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionReplace Action = "replace"
	ActionNoOp    Action = "no-op"
	ActionRead    Action = "read"
)

// AttributeChange holds the before and after values for a single resource attribute.
type AttributeChange struct {
	Before interface{}
	After  interface{}
}

// ResourceChange describes a single resource that has drifted.
type ResourceChange struct {
	// Address is the fully-qualified resource address (e.g. "aws_instance.web").
	Address string
	// Action is the planned change action.
	Action Action
	// AttributeChanges maps attribute name to its before/after values.
	// Only attributes that differ between before and after are included.
	AttributeChanges map[string]AttributeChange
}

// Plan is the parsed result of terraform plan -json output.
type Plan struct {
	// ResourceChanges lists resources with meaningful changes (non-no-op).
	ResourceChanges []ResourceChange
	// FormatVersion is the schema version reported by Terraform.
	FormatVersion string
}

// rawPlan mirrors the top-level terraform plan JSON schema.
type rawPlan struct {
	FormatVersion   string              `json:"format_version"`
	ResourceChanges []rawResourceChange `json:"resource_changes"`
}

// rawResourceChange mirrors a single resource_changes entry.
type rawResourceChange struct {
	Address string    `json:"address"`
	Change  rawChange `json:"change"`
}

// rawChange mirrors the change object within a resource_changes entry.
type rawChange struct {
	Actions []string               `json:"actions"`
	Before  map[string]interface{} `json:"before"`
	After   map[string]interface{} `json:"after"`
}

// Parse parses raw terraform plan JSON output and returns a Plan.
// Returns an error if the JSON is malformed or missing required fields.
func Parse(planJSON []byte) (*Plan, error) {
	if len(planJSON) == 0 {
		return nil, fmt.Errorf("empty plan JSON")
	}

	var raw rawPlan
	if err := json.Unmarshal(planJSON, &raw); err != nil {
		return nil, fmt.Errorf("parsing plan JSON: %w", err)
	}

	plan := &Plan{
		FormatVersion:   raw.FormatVersion,
		ResourceChanges: make([]ResourceChange, 0, len(raw.ResourceChanges)),
	}

	for _, rc := range raw.ResourceChanges {
		action := resolveAction(rc.Change.Actions)
		if action == ActionNoOp || action == ActionRead {
			continue
		}

		plan.ResourceChanges = append(plan.ResourceChanges, ResourceChange{
			Address:          rc.Address,
			Action:           action,
			AttributeChanges: diffAttributes(rc.Change.Before, rc.Change.After),
		})
	}

	return plan, nil
}

// resolveAction maps the actions array from terraform plan JSON to an Action.
// A ["delete", "create"] pair indicates a replace operation.
func resolveAction(actions []string) Action {
	if len(actions) == 0 {
		return ActionNoOp
	}
	if len(actions) >= 2 {
		return ActionReplace
	}
	switch actions[0] {
	case "create":
		return ActionCreate
	case "update":
		return ActionUpdate
	case "delete":
		return ActionDelete
	case "read":
		return ActionRead
	default:
		return ActionNoOp
	}
}

// diffAttributes returns only the attributes whose values differ between before and after.
// Attributes present in one map but absent in the other are included.
func diffAttributes(before, after map[string]interface{}) map[string]AttributeChange {
	changes := make(map[string]AttributeChange)

	keys := make(map[string]struct{}, len(before)+len(after))
	for k := range before {
		keys[k] = struct{}{}
	}
	for k := range after {
		keys[k] = struct{}{}
	}

	for k := range keys {
		bVal := before[k]
		aVal := after[k]
		if !reflect.DeepEqual(bVal, aVal) {
			changes[k] = AttributeChange{Before: bVal, After: aVal}
		}
	}

	return changes
}
