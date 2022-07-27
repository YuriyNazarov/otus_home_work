package memorystorage

import (
	"testing"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	t.Run("memory storage test", func(t *testing.T) {
		l := logger.NewLogger("debug", "STDERR")
		memStorage := New(*l)
		event := storage.Event{
			Title:   "test",
			Start:   time.Now(),
			End:     time.Now().Add(1 * time.Hour),
			OwnerID: 1,
		}
		event2 := storage.Event{
			Title:   "test",
			Start:   time.Now(),
			End:     time.Now().Add(1 * time.Hour),
			OwnerID: 1,
		}
		event3 := storage.Event{
			Title:   "test",
			Start:   time.Now(),
			End:     time.Now().Add(1 * time.Hour),
			OwnerID: 2,
		}
		event4 := storage.Event{
			Title:   "test",
			Start:   time.Now(),
			End:     time.Now().Add(1 * time.Hour),
			OwnerID: 0,
		}
		event5 := storage.Event{
			Title:   "test",
			Start:   time.Now().Add(1*time.Hour + 1*time.Second),
			End:     time.Now().Add(2 * time.Hour),
			OwnerID: 1,
		}

		err := memStorage.AddEvent(&event)
		require.NoError(t, err)
		require.NotEmpty(t, event.ID)

		err = memStorage.AddEvent(&event2)
		require.Equal(t, storage.ErrDateOverlap, err)

		err = memStorage.AddEvent(&event3)
		require.NoError(t, err)
		require.NotEmpty(t, event.ID)

		err = memStorage.AddEvent(&event4)
		require.Equal(t, storage.ErrEventDataMissing, err)

		err = memStorage.AddEvent(&event5)
		require.NoError(t, err)
		require.NotEmpty(t, event.ID)

		// Проверка получения по Id
		getEvent, err := memStorage.GetByID(event.ID)
		require.NoError(t, err)
		require.Equal(t, event, getEvent)

		_, err = memStorage.GetByID("wrong_id")
		require.Equal(t, storage.ErrEventNotFound, err)

		// Проверка получения списка
		events, err := memStorage.ListEvents(time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour+30*time.Minute), 1)
		require.NoError(t, err)
		require.Equal(t, 2, len(events))

		// Проверка удаления
		err = memStorage.DeleteEvent(event5)
		require.NoError(t, err)
		_, err = memStorage.GetByID(event5.ID)
		require.Equal(t, storage.ErrEventNotFound, err)

		// Проверка апдейта
		event.Description = "some text"
		err = memStorage.UpdateEvent(event)
		require.NoError(t, err)
		updated, err := memStorage.GetByID(event.ID)
		require.NoError(t, err)
		require.Equal(t, event.Description, updated.Description)
	})
}
