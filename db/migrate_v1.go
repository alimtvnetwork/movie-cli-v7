// migrate_v1.go — V1: Initial PascalCase schema (tables, indexes, seeds, views).
package db

// migrateV1 creates the full initial schema as a baseline migration.
func migrateV1(d *DB) error {
	if err := d.createTables(); err != nil {
		return err
	}
	if err := d.seedFileActions(); err != nil {
		return err
	}
	if err := d.seedDefaultConfig(); err != nil {
		return err
	}
	return d.createViews()
}
