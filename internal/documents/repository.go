package documents

import (
	"database/sql"
)

func GetDocuments(db *sql.DB) (map[int64]Document, error) {
	query := `SELECT id, url, title, body FROM documents LIMIT 1000`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make(map[int64]Document)

	for rows.Next() {
		var doc Document

		err := rows.Scan(
			&doc.Id,
			&doc.URL,
			&doc.Title,
			&doc.Body,
		)
		if err != nil {
			return nil, err
		}

		documents[doc.Id] = doc
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

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
