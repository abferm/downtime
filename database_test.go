package downtime_test

import (
	"testing"
	"time"

	"github.com/abferm/downtime"
	"github.com/stretchr/testify/assert"
)

func TestStoreOutbox(t *testing.T) {
	t.Run("Test Since", func(t *testing.T) {
		db, err := downtime.OpenDatabase("./test.db")
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
	})
}
