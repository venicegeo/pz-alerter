package main

import (
)


type EventDB struct {
	data map[string]Event
}

func newEventDB() *EventDB {
	db := new(EventDB)
	db.data = make(map[string]Event)
	return db
}

func (db *EventDB) write(event *Event) error {
	event.ID = newEventID()
	db.data[event.ID] = *event
	return nil
}
