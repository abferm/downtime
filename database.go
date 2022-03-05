package downtime

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

type Database struct {
	reader io.ReadSeeker
}

func NewDatabase(reader io.ReadSeeker) *Database {
	return &Database{
		reader: reader,
	}
}

func OpenDatabase(filepath string) (*Database, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return NewDatabase(file), nil
}

func (db *Database) Next() (Event, error) {
	var event Event
	err := binary.Read(db.reader, binary.BigEndian, &event)
	return event, err
}

func (db *Database) Since(after time.Time) ([]Event, error) {
	events := []Event{}
	err := db.Reset()
	if err != nil {
		return events, err
	}
	for {
		e, err := db.Next()
		if errors.Is(err, io.EOF) {
			return events, nil
		}
		if err != nil {
			return events, err
		}
		if e.When.AsTime().After(after) {
			events = append(events, e)
		}
	}
}

func (db *Database) All() ([]Event, error) {
	return db.Since(time.Time{})
}

func (db *Database) Reset() error {
	_, err := db.reader.Seek(0, io.SeekStart)
	return err
}

func (db *Database) Close() error {
	c, ok := db.reader.(io.Closer)
	if ok {
		return c.Close()
	}
	return nil
}
