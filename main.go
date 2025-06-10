package main

import (
	"bizlogger/bizlogger"
	"context"
	"log/slog"
	"os"
)

func main() {
	pgUrl := "postgres://postgres:16012006@localhost:5432/bizlogs?sslmode=disable" //from config
	techLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	bizlog, err := bizlogger.New(ctx, pgUrl)
	defer bizlog.Close()
	if err != nil {
		techLogger.Error("main.go", slog.Any("error", err))
		return
	}

	description := "Created new CA user for context test_ctx"
	bizlog.LogCreate(bizlogger.EntityUser, "admin", bizlogger.UserRoleSA, nil, "new_user", map[string]string{
		"username": "new_user",
		"role":     "CA",
		"email":    "new@example.com",
		"context":  "test_ctx",
	}, &description)

	context := "test_ctx"
	bizlog.LogUpdate(bizlogger.EntityContext, "user_ca_text_ctx", bizlogger.UserRoleCA, &context, "entity_id", map[string]string{
		"description": "Old description",
		"name":        "Old name",
	}, map[string]string{"description": "New updated description",
		"name": "Text context"})

	description2 := "Deleted user old_user with role CA"
	bizlog.LogDelete(bizlogger.EntityUser, "admin22", bizlogger.UserRoleSA, nil, "test_user", &description2)
}
