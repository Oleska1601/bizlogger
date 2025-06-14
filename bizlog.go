package bizlogger

// EventType represents the type of business event being logged.
type EventType string

// Constants for EventType values
const (
	EventTypeCreate EventType = "CREATE"
	EventTypeUpdate EventType = "UPDATE"
	EventTypeDelete EventType = "DELETE"
)

// UserRole represents the role of the user performing the action.
type UserRole string

// Constants for UserRole values
const (
	UserRoleSA UserRole = "SA" // UserRoleSA represents a System Administrator role
	UserRoleCA UserRole = "CA" // UserRoleCA represents a Client Administrator role
)

// Entity represents the type of business entity being modified.
type Entity string

// Constants for Entity values
const (
	EntityUser    Entity = "user"    // EntityUser represents user account entities.
	EntityContext Entity = "context" // EntityContext represents context entities.
)

// BizLog represents a single business log with full information.
// JSON tags ensure proper serialization for file storage and output to terminal.
type BizLog struct {
	Timestamp   string    `json:"timestamp"`             // Timestamp of when the event occurred (ISO 8601 format).
	EventType   EventType `json:"event_type"`            // EventType indicates the type of operation (CREATE/UPDATE/DELETE).
	Entity      Entity    `json:"entity"`                // Entity represents the type of business entity being modified. (user/context)
	Username    string    `json:"username"`              // Username of who performed the action.
	UserRole    UserRole  `json:"user_role"`             // UserRole represents the role of the user performing the action. (SA/CA)
	Context     *string   `json:"context,omitempty"`     // Context provides additional business context (optional).
	EntityID    string    `json:"entity_id"`             // EntityID identifies uniquely the affected entity.
	OldValue    any       `json:"old_value,omitempty"`   // OldValue contains the previous state of the entity (for UPDATE events).
	NewValue    any       `json:"new_value,omitempty"`   // NewValue contains the new state of the entity (for CREATE/UPDATE events).
	Description *string   `json:"description,omitempty"` // Description provides additional information or explanation of the event (optional).
}
