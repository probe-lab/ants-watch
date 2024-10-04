package db

// import (
// 	"context"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/probe-lab/ants-watch/db/models"
// )

// func (c *DBClient) BulkInsertRequests(ctx context.Context, requests []models.Request) error {

// 	batch := &pgx.Batch{}
// 	for _, request := range requests {

// 		batch.Queue(query, args)
// 	}

// 	results := c.Pool.SendBatch(ctx, batch)
// 	defer results.Close()

// 	for _, user := range users {
// 		_, err := results.Exec()
// 		if err != nil {
// 			var pgErr *pgconn.PgError
// 			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
// 				log.Printf("user %s already exists", user.Name)
// 				continue
// 			}

// 			return fmt.Errorf("unable to insert row: %w", err)
// 		}
// 	}

// 	return results.Close()
// }
