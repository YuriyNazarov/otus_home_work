package factory

import (
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	internalstorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/memory"
	dbstorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/sql"
)

type Storage interface {
	internalstorage.EventRepository
	internalstorage.NotificationsStorage
}

func New(config config.StorageConfig, logg logger.Logger) Storage {
	var storage Storage
	if config.MemoryStorage {
		storage = memorystorage.New(logg)
	} else {
		storage = dbstorage.New(
			logg,
			config.Database.Host,
			config.Database.Port,
			config.Database.User,
			config.Database.Password,
			config.Database.Name,
		)
	}
	return storage
}
