package sqlite3_test

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/sayerxofficial/langchaingo/tools/sqldatabase"
	_ "github.com/sayerxofficial/langchaingo/tools/sqldatabase/sqlite3"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	ctx := t.Context()
	t.Parallel()

	const dsn = `test.sqlite`
	os.Remove(dsn)
	defer os.Remove(dsn)

	// Create some example data
	tmpDB, err := sql.Open("sqlite3", dsn+"?cache=shared")
	require.NoError(t, err)

	_, err = tmpDB.Exec("CREATE TABLE `Activity` (\n  `Id` int,\n  `StringId` text,\n  `Note` text,\n  `TimeType` text,\n  `DayOfWeek` text,\n  `Year` text,\n  `Month` text,\n  `Day` text,\n  `Hour` text,\n  `Minute` text,\n  `Second` text,\n  `Duration` int\n) ") //nolint:lll
	require.NoError(t, err)
	_, err = tmpDB.Exec("CREATE TABLE `Activity1` (\n  `Id` int,\n  `StringId` text,\n  `Note` text,\n  `TimeType` text,\n  `DayOfWeek` text,\n  `Year` text,\n  `Month` text,\n  `Day` text,\n  `Hour` text,\n  `Minute` text,\n  `Second` text,\n  `Duration` int\n)  ") //nolint:lll
	require.NoError(t, err)
	_, err = tmpDB.Exec("CREATE TABLE `Activity2` (\n  `Id` int,\n  `StringId` text,\n  `Note` text,\n  `TimeType` text,\n  `DayOfWeek` text,\n  `Year` text,\n  `Month` text,\n  `Day` text,\n  `Hour` text,\n  `Minute` text,\n  `Second` text,\n  `Duration` int\n)  ") //nolint:lll
	require.NoError(t, err)
	tmpDB.Close()

	db, err := sqldatabase.NewSQLDatabaseWithDSN("sqlite3", dsn, nil)
	require.NoError(t, err)
	defer db.Close()

	tbs := db.TableNames()
	require.Len(t, tbs, 3)

	desc, err := db.TableInfo(ctx, tbs)
	require.NoError(t, err)

	desc = strings.TrimSpace(desc)
	require.Equal(t, 0, strings.Index(desc, "CREATE TABLE")) //nolint:stylecheck
	require.True(t, strings.Contains(desc, "Activity"))      //nolint:stylecheck
	require.True(t, strings.Contains(desc, "Activity1"))     //nolint:stylecheck
	require.True(t, strings.Contains(desc, "Activity2"))     //nolint:stylecheck

	for _, tableName := range tbs {
		_, err = db.Query(ctx, fmt.Sprintf("SELECT * from %s LIMIT 1", tableName))
		/* exclude no row error,
		since we only need to check if db.Query function can perform query correctly*/
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		require.NoError(t, err)
	}
}
