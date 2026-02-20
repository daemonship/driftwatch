// Package parser parses the JSON output from terraform plan -json.
// Requires Terraform >= 1.0.0.
package parser

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

// Parse parses raw terraform plan JSON output and returns a Plan.
// Returns an error if the JSON is malformed or missing required fields.
func Parse(planJSON []byte) (*Plan, error) {
	// TODO: implement in Task 3
	return nil, nil
}
