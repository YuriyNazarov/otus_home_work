package sqlstorage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	_ "github.com/lib/pq" // Postgres driver
	goose "github.com/pressly/goose/v3"
	uuid "github.com/satori/go.uuid"
)

type Storage struct {
	db  *sql.DB
	log logger.Logger
}

func (s *Storage) AddEvent(event *storage.Event) error {
	if !event.IsRequiredFilled() {
		return storage.ErrEventDataMissing
	}
	exEvents, _ := s.ListEvents(event.Start, event.End, event.OwnerID)
	if len(exEvents) > 0 {
		return storage.ErrDateOverlap
	}
	event.ID = uuid.NewV4().String()
	query := "insert into events (id, title, description, owner_id, start, \"end\", remind_before)  values" +
		" ($1, $2, $3, $4, $5, $6, $7)"
	_, err := s.db.Exec(
		query,
		event.ID,
		event.Title,
		event.Description,
		event.OwnerID,
		event.Start,
		event.End,
		event.RemindBefore.Nanoseconds(),
	)
	if err != nil {
		s.log.Error(fmt.Sprintf("adding event failed: %s", err))
		return storage.ErrDBOperationFail
	}
	s.log.Debug("stored event id=" + event.ID)
	return nil
}

func (s *Storage) UpdateEvent(event storage.Event) error {
	if !event.IsRequiredFilled() {
		return storage.ErrEventDataMissing
	}
	events, _ := s.listEventsExclude(event.Start, event.End, event.OwnerID, event.ID)
	if len(events) > 0 {
		return storage.ErrDateOverlap
	}
	query := "update events set" +
		" title = $1," +
		" description = $2," +
		" owner_id = $3," +
		" start = $4," +
		" \"end\" = $5," +
		" remind_before = $6," +
		" remind_sent = $7," +
		" remind_received= $8" +
		" where id = $9"
	_, err := s.db.Exec(
		query,
		event.Title,
		event.Description,
		event.OwnerID,
		event.Start,
		event.End,
		event.RemindBefore.Nanoseconds(),
		event.RemindSent,
		event.RemindReceived,
		event.ID,
	)
	if err != nil {
		s.log.Error(fmt.Sprintf("updating event failed: %s", err))
		return storage.ErrDBOperationFail
	}
	s.log.Debug("updated event id=" + event.ID)
	return nil
}

func (s *Storage) DeleteEvent(event storage.Event) error {
	query := "delete from events where id = $1"
	_, err := s.db.Exec(query, event.ID)
	if err != nil {
		s.log.Error(fmt.Sprintf("deleting event failed: %s", err))
		return storage.ErrDBOperationFail
	}
	s.log.Debug("deleted event id=" + event.ID)
	return nil
}

func (s *Storage) ListEvents(from time.Time, to time.Time, ownerID int) ([]storage.Event, error) {
	return s.listEventsExclude(from, to, ownerID, "")
}

func (s *Storage) listEventsExclude(
	from time.Time,
	to time.Time,
	ownerID int,
	excludeID string,
) ([]storage.Event, error) {
	query := "select id, title, description, owner_id, start, \"end\", remind_before, remind_sent, remind_received" +
		" from events where ((\"start\" >= $1 and \"start\" <= $2) or (\"end\" >= $1 and \"end\" <= $2)) and id != $3"
	var (
		err  error
		rows *sql.Rows
	)
	if ownerID != 0 {
		query += " and owner_id = $4"
		rows, err = s.db.Query(query, from, to, excludeID, ownerID)
	} else {
		rows, err = s.db.Query(query, from, to, excludeID)
	}

	var events []storage.Event
	if err != nil {
		return events, err
	}
	defer rows.Close()
	for rows.Next() {
		var event storage.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.OwnerID,
			&event.Start,
			&event.End,
			&event.IntRemindBefore,
			&event.RemindSent,
			&event.RemindReceived,
		)
		if err != nil {
			s.log.Error(fmt.Sprintf("err on scanning rows: %s", err))
			continue
		}
		event.RemindBefore = time.Duration(event.IntRemindBefore)
		events = append(events, event)
	}
	if rows.Err() != nil {
		s.log.Error(fmt.Sprintf("err on scanning rows: %s", err))
	}
	return events, nil
}

func (s *Storage) GetByID(id string) (storage.Event, error) {
	query := "select id, title, description, owner_id, start, \"end\", remind_before, remind_sent, remind_received" +
		" from events where id = $1"
	var event storage.Event
	err := s.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.OwnerID,
		&event.Start,
		&event.End,
		&event.IntRemindBefore,
		&event.RemindSent,
		&event.RemindReceived,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return event, storage.ErrEventNotFound
		}
		s.log.Error(fmt.Sprintf("finding event failed: %s", err))
		return event, storage.ErrDBOperationFail
	}
	event.RemindBefore = time.Duration(event.IntRemindBefore)
	return event, nil
}

func New(l logger.Logger, host string, port int, user, password, dbName string) *Storage {
	storageInstance := Storage{
		log: l,
	}
	err := storageInstance.Connect(host, port, user, password, dbName)
	if err != nil {
		l.Error(fmt.Sprintf("error occupied on connecting to DB: %s", err))
	}
	err = storageInstance.migrate()
	if err != nil {
		l.Error(fmt.Sprintf("error occupied on migrating: %s", err))
		return &Storage{}
	}
	return &storageInstance
}

func (s *Storage) Connect(host string, port int, user, password, dbName string) error {
	s.log.Debug("connect DB")
	connection := "postgres://" + user + ":" + password + "@" +
		host + ":" + strconv.Itoa(port) + "/" + dbName + "?sslmode=disable"

	db, err := sql.Open("postgres", connection)
	if err != nil {
		s.log.Error(fmt.Sprintf("failed on connecting to db: %s", err))
		return storage.ErrConnFailed
	}
	err = db.Ping()
	if err != nil {
		s.log.Error(fmt.Sprintf("db healthcheck failed: %s", err))
		return storage.ErrConnFailed
	}
	s.db = db
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) migrate() error {
	pwd, _ := os.Getwd()
	migrPath := filepath.Join(pwd, "../migrations")
	return goose.Up(s.db, migrPath)
}
