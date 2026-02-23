package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/connectors/database"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (a *Agent) toolQueryDatabase(ctx context.Context, userID uuid.UUID, input map[string]any) (string, error) {
	sql, _ := input["sql"].(string)
	if sql == "" {
		return "", fmt.Errorf("query_database: sql parameter is required")
	}

	connIDStr, _ := input["connection_id"].(string)
	if connIDStr == "" {
		return "", fmt.Errorf("query_database: connection_id parameter is required")
	}

	connID, err := uuid.Parse(connIDStr)
	if err != nil {
		return "", fmt.Errorf("query_database: invalid connection_id: %w", err)
	}

	// Fetch the connection — user-scoped to prevent unauthorized access.
	conn, err := a.queries.GetConnectionByUser(ctx, db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		return "", fmt.Errorf("query_database: connection not found or not owned by user")
	}

	// Validate connection type.
	if conn.Type != "database" && conn.Type != "postgres" {
		return "", fmt.Errorf("query_database: connection type %q is not a database", conn.Type)
	}

	// Create connector, connect, query, close.
	pg, err := database.New(conn.Config)
	if err != nil {
		return "", fmt.Errorf("query_database: %w", err)
	}

	if err := pg.Connect(ctx); err != nil {
		return "", fmt.Errorf("query_database: %w", err)
	}
	defer pg.Close(ctx)

	rows, err := pg.Query(ctx, sql)
	if err != nil {
		return "", fmt.Errorf("query_database: %w", err)
	}

	result, err := json.Marshal(map[string]any{
		"rows":      rows,
		"row_count": len(rows),
	})
	if err != nil {
		return "", fmt.Errorf("query_database: marshal: %w", err)
	}
	return string(result), nil
}
