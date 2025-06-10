package bizlogger

type EventType string

const (
	EventTypeCreate EventType = "CREATE"
	EventTypeUpdate EventType = "UPDATE"
	EventTypeDelete EventType = "DELETE"
)

type UserRole string

const (
	UserRoleSA UserRole = "SA"
	UserRoleCA UserRole = "CA"
)

type Entity string

const (
	EntityUser    Entity = "user"
	EntityContext Entity = "context"
)
