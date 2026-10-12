package admingt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
)

// The admin works straight off the live Postgres catalog instead of a hand-kept model registry,
// so every table (present and future) is automatically browsable and editable.

type column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
	Primary  bool   `json:"primary"`
	// Hidden columns (password hashes, tokens, PESEL ciphertext...) are never sent to the browser.
	Hidden bool `json:"hidden"`
}

type tableInfo struct {
	Name    string   `json:"name"`
	Columns []column `json:"columns"`
	PK      []string `json:"pk"`
}

var (
	schemaMu    sync.Mutex
	schemaCache = map[string]*tableInfo{}
	schemaAt    time.Time
	tableNames  []string
)

var hiddenRe = regexp.MustCompile(`(?i)(password|passwd|hash|secret|token|pesel|salt|otp)`)

func loadSchema(ctx context.Context) error {
	schemaMu.Lock()
	defer schemaMu.Unlock()
	if time.Since(schemaAt) < 30*time.Second && len(schemaCache) > 0 {
		return nil
	}
	type nameRow struct{ TableName string }
	names, err := gormdb.CustomSelect[nameRow](ctx, `select table_name from information_schema.tables
 where table_schema='public' and table_type='BASE TABLE' order by table_name`)
	if err != nil {
		return err
	}
	type colRow struct {
		TableName, ColumnName, DataType, IsNullable string
		ColumnDefault                               *string
	}
	cols, err := gormdb.CustomSelect[colRow](ctx, `select table_name, column_name, data_type, is_nullable, column_default
 from information_schema.columns where table_schema='public' order by table_name, ordinal_position`)
	if err != nil {
		return err
	}
	type pkRow struct{ TableName, ColumnName string }
	pks, err := gormdb.CustomSelect[pkRow](ctx, `select tc.table_name, kcu.column_name
 from information_schema.table_constraints tc
 join information_schema.key_column_usage kcu on kcu.constraint_name = tc.constraint_name and kcu.table_schema = tc.table_schema
 where tc.table_schema='public' and tc.constraint_type='PRIMARY KEY' order by kcu.ordinal_position`)
	if err != nil {
		return err
	}
	next := map[string]*tableInfo{}
	tableNames = tableNames[:0]
	for _, n := range names {
		next[n.TableName] = &tableInfo{Name: n.TableName, Columns: []column{}, PK: []string{}}
		tableNames = append(tableNames, n.TableName)
	}
	for _, c := range cols {
		t := next[c.TableName]
		if t == nil {
			continue
		}
		col := column{Name: c.ColumnName, Type: c.DataType, Nullable: c.IsNullable == "YES", Hidden: hiddenRe.MatchString(c.ColumnName)}
		if c.ColumnDefault != nil {
			col.Default = *c.ColumnDefault
		}
		t.Columns = append(t.Columns, col)
	}
	for _, p := range pks {
		t := next[p.TableName]
		if t == nil {
			continue
		}
		t.PK = append(t.PK, p.ColumnName)
		for i := range t.Columns {
			if t.Columns[i].Name == p.ColumnName {
				t.Columns[i].Primary = true
			}
		}
	}
	schemaCache, schemaAt = next, time.Now()
	return nil
}

func getTable(ctx context.Context, name string) (*tableInfo, error) {
	if err := loadSchema(ctx); err != nil {
		return nil, err
	}
	schemaMu.Lock()
	defer schemaMu.Unlock()
	t := schemaCache[name]
	if t == nil {
		return nil, fmt.Errorf("nie ma tabeli %q", name)
	}
	return t, nil
}

func (t *tableInfo) col(name string) *column {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}
	return nil
}

// q quotes an identifier that has already been validated against the catalog.
func q(ident string) string { return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"` }

func isTextual(c column) bool {
	switch c.Type {
	case "text", "character varying", "character", "uuid":
		return true
	}
	return false
}

func isIntType(t string) bool { return t == "integer" || t == "bigint" || t == "smallint" }
