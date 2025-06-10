package bizlogger

type BizLog struct {
	Timestamp   string      `json:"timestamp"`
	EventType   EventType   `json:"event_type"`
	Entity      Entity      `json:"entity"`
	Username    string      `json:"username"`
	UserRole    UserRole    `json:"user_role"`
	Context     *string     `json:"context,omitempty"`
	EntityID    string      `json:"entity_id"`
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	Description *string     `json:"description,omitempty"`
}
