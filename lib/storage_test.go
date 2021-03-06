package lib

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DBFilepath = "/tmp/spiderwoman.db"
)

func TestCreateDBIfNotExists(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)
	_, err := os.Stat(DBFilepath)
	assert.Equal(t, false, os.IsNotExist(err))
}

func TestCheckMonitorTable(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)

	db, err := sql.Open("sqlite3", DBFilepath)
	assert.Equal(t, nil, err)
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE name='monitor';")
	assert.Equal(t, nil, err)
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		assert.Equal(t, nil, err)
		assert.Equal(t, "monitor", name)
	}

	err = rows.Err()
	assert.Equal(t, nil, err)
}

func TestCheckStatusTable(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)

	db, err := sql.Open("sqlite3", DBFilepath)
	assert.Equal(t, nil, err)
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE name='status';")
	assert.Equal(t, nil, err)
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		assert.Equal(t, nil, err)
		assert.Equal(t, "status", name)
	}

	err = rows.Err()
	assert.Equal(t, nil, err)
}

func TestSaveRecordToMonitor(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)

	sourceHost := "http://a"
	externalLink := "http://b/1"
	count := 800
	externalHost := "b"
	res := SaveRecordToMonitor(DBFilepath, sourceHost, externalLink, count, externalHost)
	assert.Equal(t, true, res)

	db, err := sql.Open("sqlite3", DBFilepath)
	assert.Equal(t, nil, err)
	defer db.Close()

	rows, err := db.Query("SELECT source_host, created FROM monitor;")
	assert.Equal(t, nil, err)
	defer rows.Close()

	k := 0
	for rows.Next() {
		k++
		var sourceHost string
		var created string
		err = rows.Scan(&sourceHost, &created)
		assert.Equal(t, nil, err)
		assert.Equal(t, "http://a", sourceHost)
	}
	assert.Equal(t, 1, k)
}

func TestSaveRecordToMonitor_BadExternalLink(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)

	sourceHost := "http://a"
	externalLink := "http://somebaddomain.com/'.show_site_name($line['url'],100).'/"
	count := 800
	externalHost := "b"
	res := SaveRecordToMonitor(DBFilepath, sourceHost, externalLink, count, externalHost)
	assert.Equal(t, true, res)

	db, err := sql.Open("sqlite3", DBFilepath)
	assert.Equal(t, nil, err)
	defer db.Close()

	rows, err := db.Query("SELECT external_link FROM monitor;")
	assert.Equal(t, nil, err)
	defer rows.Close()

	k := 0
	for rows.Next() {
		k++
		var fetchedExternalLink string
		err = rows.Scan(&fetchedExternalLink)
		assert.Equal(t, nil, err)
		assert.Equal(t, externalLink, fetchedExternalLink)
	}
	assert.Equal(t, 1, k)
}

func TestGetAllDataFromSqlite_MapToStruct(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)
	err := SaveHostType(DBFilepath, "host1", "type1")
	assert.NoError(t, err)
	err = SaveHostType(DBFilepath, "host2", "type2")
	assert.NoError(t, err)

	for i := int(0); i < 10; i++ {
		sourceHost := "host1"
		externalLink := "http://b/1?" + strconv.Itoa(i)
		count := 800+i
		externalHost := "b"
		_ = SaveRecordToMonitor(DBFilepath, sourceHost, externalLink, count, externalHost)
	}

	sourceHost := "host2"
	externalLink := "http://b/1?10"
	count := 810
	externalHost := "host1"
	_ = SaveRecordToMonitor(DBFilepath, sourceHost, externalLink, count, externalHost)

	sourceHost = "host3"
	externalLink = "http://b/1?10"
	count = 810
	externalHost = "host3"
	_ = SaveRecordToMonitor(DBFilepath, sourceHost, externalLink, count, externalHost)

	monitors, err := GetAllDataFromMonitor(DBFilepath, 9)
	assert.NoError(t, err)
	assert.Equal(t, 12, len(monitors))
	assert.Equal(t, "http://b/1?0", monitors[0].ExternalLink)
	assert.Equal(t, "http://b/1?9", monitors[9].ExternalLink)

	assert.Equal(t, "type1", monitors[0].SourceHostType)
	assert.Equal(t, "H", monitors[0].ExternalHostType)
	assert.Equal(t, "type2", monitors[10].SourceHostType)
	assert.Equal(t, "type1", monitors[10].ExternalHostType)
	assert.Equal(t, "H", monitors[11].SourceHostType)
	assert.Equal(t, "H", monitors[11].ExternalHostType)

	//for _, m := range monitors {
	//	log.Printf("[%v] [%v] %v %v %v", m.SourceHostType, m.ExternalHostType, m.Created, m.ExternalHost, m.SourceHost)
	//}
}

func TestParseSqliteDate(t *testing.T) {
	sqliteDate := "2016-12-13T07:17:23Z"
	parsedTime, _ := ParseSqliteDate(sqliteDate)
	assert.Equal(t, "2016", strconv.Itoa(parsedTime.Year()))
	assert.Equal(t, "12", strconv.Itoa(int(parsedTime.Month())))
	assert.Equal(t, "13", strconv.Itoa(parsedTime.Day()))
}

func TestGetDataFromMonitor__ByDays(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)
	err := SaveHostType(DBFilepath, "host1", "type1")
	assert.NoError(t, err)

	for i := int(0); i < 100; i++ {
		sourceHost := "http://a"
		externalLink := "http://b/1?" + strconv.Itoa(i)
		count := 800+i
		externalHost := "b"
		_ = SaveRecordToMonitor(DBFilepath, sourceHost, externalLink, count, externalHost)
	}

	dates, err := GetAllDaysFromMonitor(DBFilepath)
	assert.NoError(t, err)
	assert.Equal(t, len(dates), 1)

	monitors, err := GetAllDataFromMonitorByDay(DBFilepath, dates[0])
	assert.NoError(t, err)
	assert.Equal(t, len(monitors), 100)

	monitors, err = GetAllDataFromMonitorByDay(DBFilepath, "2122-01-01")
	assert.NoError(t, err)
	assert.Equal(t, len(monitors), 0)

	monitors, err = GetAllDataFromMonitorByDay(DBFilepath, "1812-01-01")
	assert.NoError(t, err)
	assert.Equal(t, len(monitors), 0)
}

func TestCrawlStatus(t *testing.T) {
	CreateDBIfNotExists(DBFilepath)

	SetCrawlStatus(DBFilepath, "Crawling...")
	s1, _ := GetCrawlStatus(DBFilepath)
	assert.Equal(t, "Crawling...", s1)

	SetCrawlStatus(DBFilepath, "Crawl done")
	GetCrawlStatus(DBFilepath)
	s2, _ := GetCrawlStatus(DBFilepath)
	assert.Equal(t, "Crawl done", s2)

	CreateDBIfNotExists(DBFilepath)
	SetCrawlStatus(DBFilepath, "Crawling...")
	s3, _ := GetCrawlStatus(DBFilepath)
	assert.Equal(t, "Crawling...", s3)
}

func TestSaveHostType(t *testing.T) {
	os.Remove(DBFilepath)
	CreateDBIfNotExists(DBFilepath)
	err := SaveHostType(DBFilepath, "host1", "type1")
	assert.NoError(t, err)
	err = SaveHostType(DBFilepath, "host1", "type1")
	assert.Error(t, err)
	err = SaveHostType(DBFilepath, "host1", "type2")
	assert.Error(t, err)
	err = SaveHostType(DBFilepath, "host2", "type1")
	assert.NoError(t, err)
}

func TestDeleteTypesTable(t *testing.T) {
	err := DeleteTypesTable(DBFilepath)
	assert.NoError(t, err)
}