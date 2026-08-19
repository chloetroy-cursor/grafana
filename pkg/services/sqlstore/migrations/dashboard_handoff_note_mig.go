package migrations

import (
	. "github.com/grafana/grafana/pkg/services/sqlstore/migrator"
)

func addDashboardHandoffNoteMigrations(mg *Migrator) {
	note := Table{
		Name: "dashboard_handoff_note",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "dashboard_uid", Type: DB_NVarchar, Length: 40, Nullable: false},
			{Name: "org_id", Type: DB_BigInt, Nullable: false},
			{Name: "author_id", Type: DB_BigInt, Nullable: false},
			{Name: "text", Type: DB_Text, Nullable: false},
			{Name: "created_at", Type: DB_DateTime, Nullable: false},
		},
		Indices: []*Index{
			{Cols: []string{"org_id", "dashboard_uid", "created_at"}},
		},
	}

	mention := Table{
		Name: "dashboard_handoff_note_mention",
		Columns: []*Column{
			{Name: "id", Type: DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "note_id", Type: DB_BigInt, Nullable: false},
			{Name: "value", Type: DB_NVarchar, Length: 190, Nullable: false},
			{Name: "user_id", Type: DB_BigInt, Nullable: true},
		},
		Indices: []*Index{
			{Cols: []string{"note_id"}},
		},
	}

	mg.AddMigration("create dashboard handoff note table", NewAddTableMigration(note))
	mg.AddMigration("add dashboard handoff note dashboard index", NewAddIndexMigration(note, note.Indices[0]))
	mg.AddMigration("create dashboard handoff note mention table", NewAddTableMigration(mention))
	mg.AddMigration("add dashboard handoff note mention index", NewAddIndexMigration(mention, mention.Indices[0]))
}
