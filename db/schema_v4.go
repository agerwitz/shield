package db

type v4Schema struct{}

func (s v4Schema) Deploy(db *DB) error {
	var err error

	err = db.Exec(`ALTER TABLE archives ADD COLUMN encryption_mode VARCHAR(50)`)
	if err != nil {
		return err
	}

	return nil
}
