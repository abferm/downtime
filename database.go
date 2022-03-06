package downtime

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

func NewDatabaseWriter(writer io.Writer) *DatabaseWriter {
	return &DatabaseWriter{
		writer: writer,
	}
}

type DatabaseWriter struct {
	writer io.Writer
}

func (db *DatabaseWriter) Append(event Event) error {
	err := binary.Write(db.writer, binary.BigEndian, event)
	if err != nil {
		return err
	}
	sync, ok := db.writer.(syncer)
	if ok {
		return sync.Sync()
	}
	return nil
}

func (db *DatabaseWriter) Close() error {
	c, ok := db.writer.(io.Closer)
	if ok {
		return c.Close()
	}
	return nil
}

type DatabaseReader struct {
	reader io.ReadSeeker
}

func NewDatabaseReader(reader io.ReadSeeker) *DatabaseReader {
	return &DatabaseReader{
		reader: reader,
	}
}

func OpenDatabaseReader(filepath string) (*DatabaseReader, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return NewDatabaseReader(file), nil
}

func (db *DatabaseReader) Next() (Event, error) {
	var event Event
	err := binary.Read(db.reader, binary.BigEndian, &event)
	return event, err
}

func (db *DatabaseReader) Since(after time.Time) ([]Event, error) {
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

func (db *DatabaseReader) All() ([]Event, error) {
	return db.Since(time.Time{})
}

func (db *DatabaseReader) Reset() error {
	_, err := db.reader.Seek(0, io.SeekStart)
	return err
}

func (db *DatabaseReader) Close() error {
	c, ok := db.reader.(io.Closer)
	if ok {
		return c.Close()
	}
	return nil
}

type syncer interface {
	Sync() error
}
