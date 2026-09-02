package documents

import "database/sql"

func SaveDocument(db *sql.DB, document *T_Document) error {
	query := `
        INSERT INTO documents (
            url,
            title,
            body,
            domain
        )
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (url)
        DO UPDATE SET
            title = EXCLUDED.title,
            body = EXCLUDED.body
    `

	_, err := db.Exec(
		query,
		document.URL,
		document.Title,
		document.Body,
		document.Domain,
	)

	return err
}
