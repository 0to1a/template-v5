package database

import "github.com/jackc/pgx/v5/pgtype"

// ParsePublicUUID converts a public-facing UUID string into the pgtype.UUID
// used for querying by primary/foreign key columns. Domain repositories
// share this so a malformed public UUID is always handled the same way
// (callers map the error to their own "not found" sentinel).
func ParsePublicUUID(publicUUID string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(publicUUID); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}
