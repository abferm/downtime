package downtime_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/abferm/downtime"
	"github.com/stretchr/testify/assert"
)

func TestReaderSince(t *testing.T) {
	db, err := downtime.OpenDatabaseReader("./test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	events, err := db.Since(time.Unix(1633484567, 0))
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, events, 3)

	for _, event := range events {
		if event.When.AsTime().After(time.Unix(1633484567, 0)) == false {
			assert.True(t, time.Unix(1633484567, 0).After(event.When.AsTime()))
		}
	}
}

func TestWriterAppend(t *testing.T) {
	r, err := downtime.OpenDatabaseReader("./test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	expectedEvents, err := r.All()
	if err != nil {
		t.Fatal(err)
	}

	buff := bytes.NewBuffer([]byte{})
	w := downtime.NewDatabaseWriter(buff)
	defer w.Close()
	for _, event := range expectedEvents {
		err = w.Append(event)
		if err != nil {
			t.Fatal(err)
		}
	}

	r2 := downtime.NewDatabaseReader(bytes.NewReader(buff.Bytes()))
	actualEvents, err := r2.All()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedEvents, actualEvents)
}
