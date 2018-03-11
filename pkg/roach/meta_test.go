package roach_test

import (
	"flag"
	"testing"
	"github.com/tomogoma/imagems/pkg/roach"
	"github.com/tomogoma/imagems/pkg/model"
)


func init() {
	flag.Parse()
}

func TestDB_SaveMeta(t *testing.T) {
	conf, tearDown := setup(t)
	defer tearDown()

	d := roach.New(getOpts(conf)...)
	meta := model.ImageMeta{UserID: "1234"}
	ID, err := d.SaveMeta(meta)
	if err != nil {
		t.Fatalf("db.SaveMeta(): %v", err)
	}
	if ID < 1 {
		t.Fatal("meta ID not propagated upwards to caller")
	}
}

func TestDB_DeleteMeta(t *testing.T) {
	conf, tearDown := setup(t)
	defer tearDown()

	d := roach.New(getOpts(conf)...)
	meta := model.ImageMeta{UserID: "1234"}
	ID, err := d.SaveMeta(meta)
	if err != nil {
		t.Fatalf("db.SaveMeta(): %v", err)
	}
	if err := d.DeleteMeta(ID); err != nil {
		t.Fatalf("db.DeleteMeta(): %v", err)
	}
}
