package handler

// orEmptyStrings normalizes a nil slice to a non-nil empty slice so a NOT NULL
// Postgres array column receives '{}' rather than NULL when the client omits it.
func orEmptyStrings(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}
