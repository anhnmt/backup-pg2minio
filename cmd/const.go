package cmd

const (
	PSQL      = "psql"
	PgDump    = "pg_dump"
	PgRestore = "pg_restore"
	Gzip      = "gzip"
	Gunzip    = "gunzip"
)

const (
	Alias      = "minio"
	PgDumpFile = "pg_dump.sql.gz"
	// PgDumpFileCustom is the output file for custom format
	PgDumpFileCustom = "pg_dump.custom.gz"
	// PgDumpFileDirectory is the output directory for directory format
	PgDumpFileDirectory = "pg_dump.backup"
	// PgDumpFilePlain is the output file for plain format
	PgDumpFilePlain = "pg_dump.sql.gz"
)
