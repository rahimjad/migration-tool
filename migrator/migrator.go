package migrator

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"../postgres"
)

type migrationRecord struct {
	version   int64
	succeeded bool
	createdAt int64
	direction string
}

type migrationFileMetaData struct {
	validFile bool
	version   int64
	direction string
	path      string
}

const migrationDir = "migrations"

// MigrateUp it does stuff
func MigrateUp() {
	createMigraitonTable()

	m := getLatestMigration()

	if m.version == 0 && m.direction == "" {
		log.Output(1, "No migrations exist. Running all migrations from default folder.")
	}
	runMigrations(m.version)
}

func createMigraitonTable() {
	postgres.Exec(`
		BEGIN;
		CREATE TABLE IF NOT EXISTS migrations
		(
			version bigserial not null,
			succeeded boolean not null default false,
			createdAt bigserial not null,
			direction varchar not null
		);
		CREATE index index_migrations_on_versions on migrations (version);
		COMMIT;
	`)
}

func getLatestMigration() migrationRecord {
	row := postgres.QueryRow(`
		SELECT * FROM migrations
		ORDER BY version DESC
		LIMIT 1;
	`)

	var m migrationRecord

	err := row.Scan(&m.version, &m.succeeded, &m.createdAt, &m.direction)

	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return m
}

func runMigrations(version int64) {
	pwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Could not get root")
	}

	mFilePath := filepath.Join(pwd, migrationDir)

	err = filepath.Walk(mFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Output(1, fmt.Sprintf("Prevent panic by handling failure accessing a path %q: %v\n", path, err))
			return err
		}

		if info.Name() == migrationDir {
			return err
		}

		if info.IsDir() {
			fmt.Printf("Skipping a dir without errors: %+v \n", path)
			return filepath.SkipDir
		}

		migrationInfo := extractMigrationInfo(info.Name(), path)

		if migrationInfo.validFile == false {
			log.Output(1, fmt.Sprintf("Invalid migration file found at %q\n", path))
			return err
		}

		if migrationInfo.version <= version {
			log.Output(1, fmt.Sprintf("Skipping migration version: %d", migrationInfo.version))
			return err
		}

		runMigrationWithMetaData(migrationInfo)

		return err
	})

	if err != nil {
		panic(err)
	}
}

func extractMigrationInfo(fileName string, path string) migrationFileMetaData {
	var m migrationFileMetaData

	m.path = path

	match, err := regexp.MatchString("^\\d*_\\w*.((up)|(down)).sql$", fileName)

	if err != nil {
		log.Panic(err)
	}

	if match == false {
		return m
	}

	m.validFile = match

	versionRegex := regexp.MustCompile("^\\d")
	versionStr := versionRegex.FindString(fileName)

	version, err := strconv.Atoi(versionStr)

	m.version = int64(version)

	directionRegex := regexp.MustCompile("(up|down)")
	direction := directionRegex.FindString(fileName)

	m.direction = direction

	return m
}

func runMigrationWithMetaData(m migrationFileMetaData) {
	buf, err := ioutil.ReadFile(m.path)

	if err != nil {
		panic(err)
	}

	sql := string(buf)

	sql = fmt.Sprintf("BEGIN;\n%s\nCOMMIT;", sql)

	sqlArr := strings.Split(sql, ";")

	for _, query := range sqlArr {
		_, err = postgres.Exec(query)
	}

	success := err == nil

	if success {
		log.Output(1, fmt.Sprintf("Migration Version: %v COMPLETED", m.version))
	} else {
		log.Output(1, fmt.Sprintf("Migration Version: %v FAILED", m.version))
		log.Fatal(err)
	}

	insertSQL := fmt.Sprintf(`
		INSERT INTO migrations (version, succeeded, direction, createdAt)
		VALUES (%d, %t, '%s', %d)
	`,
		m.version,
		success,
		m.direction,
		makeTimestamp(),
	)

	_, err = postgres.Exec(insertSQL)

	if err != nil {
		log.Fatal(err)
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano()
}
