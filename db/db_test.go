package db_test

import (
	"flag"
	"github.com/tomogoma/go-commons/config"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/imagems/db"
	"github.com/tomogoma/imagems/model"
	"testing"
)

type ConfigMock struct {
	Database cockroach.DSN `yaml:"database,omitempty"`
}

var confPath = flag.String("conf", "/etc/imagems/imagems.conf.yaml", "/path/to/imagems.conf.yaml")
var validConf *ConfigMock

func init() {
	flag.Parse()
}

func TestNew(t *testing.T) {
	setUp(t)
	defer TearDown(t)
	d, err := db.New(validConf.Database)
	if err != nil {
		t.Fatalf("db.New(): %v", err)
	}
	if d == nil {
		t.Fatal("Got a nil DB")
	}
}

func TestNew_nilDsnF(t *testing.T) {
	_, err := db.New(nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestDB_SaveMeta(t *testing.T) {
	setUp(t)
	defer TearDown(t)
	d, err := db.New(validConf.Database)
	if err != nil {
		t.Fatalf("db.New(): %v", err)
	}
	meta := &model.ImageMeta{UserID: 1234}
	if err := d.SaveMeta(meta); err != nil {
		t.Fatalf("db.SaveMeta(): %v", err)
	}
	if meta.ID < 1 {
		t.Fatal("meta ID not propagated upwards to caller")
	}
}

func TestDB_DeleteMeta(t *testing.T) {
	setUp(t)
	defer TearDown(t)
	d, err := db.New(validConf.Database)
	if err != nil {
		t.Fatalf("db.New(): %v", err)
	}
	meta := &model.ImageMeta{UserID: 1234}
	if err := d.SaveMeta(meta); err != nil {
		t.Fatalf("db.SaveMeta(): %v", err)
	}
	if err := d.DeleteMeta(meta.ID); err != nil {
		t.Fatalf("db.DeleteMeta(): %v", err)
	}
}

func setUp(t *testing.T) {
	validConf = &ConfigMock{}
	if err := config.ReadYamlConfig(*confPath, validConf); err != nil {
		t.Fatalf("error setting up (reading config file): %v", err)
	}
	validConf.Database.DB = validConf.Database.DB + "_test"
}

func TearDown(t *testing.T) {
	d, err := cockroach.DBConn(validConf.Database)
	if err != nil {
		t.Logf("Error tearing down (connecting to db): %v", err)
		return
	}
	q := `DROP DATABASE IF EXISTS ` + validConf.Database.DB
	if _, err := d.Exec(q); err != nil {
		t.Logf("Error tearing down (deleting db): %v", err)
		return
	}
}
