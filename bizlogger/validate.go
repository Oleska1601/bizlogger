package bizlogger

import "errors"

// Общая валидация

// Проверка правил контекста
func validateContextRules(role UserRole, context *string) error {
	if role == UserRoleSA && context != nil {
		return errors.New("SA cannot have context")
	}

	if role == UserRoleCA && context == nil {
		return errors.New("CA must have context")
	}

	return nil
}
