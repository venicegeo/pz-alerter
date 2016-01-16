package main

import (
	"strconv"
)

//---------------------------------------------------------------------------

var conditionID = 1
func newConditionID() string {
	s := strconv.Itoa(conditionID)
	conditionID++
	return s
}

type Condition struct {
	ID        string `json:"id"`
	Name      string `json:"name" binding:"required"`
	Condition string `json:"condition" binding:"required"`
}

type ConditionDB struct {
	data map[string]Condition
}

func newConditionDB() *ConditionDB {
	db := new(ConditionDB)
	db.data = make(map[string]Condition)
	return db
}

func (db *ConditionDB) write(condition *Condition) error {
	condition.ID = newConditionID()
	db.data[condition.ID] = *condition
	return nil
}

func (db *ConditionDB) update(condition *Condition) bool {
	_, ok := db.data[condition.ID]
	if ok {
		db.data[condition.ID] = *condition
		return true
	}
	return false
}

func (db *ConditionDB) readByID(id string) *Condition {
	v, ok := db.data[id]
	if !ok {
		return nil
	}
	return &v
}

func (db *ConditionDB) deleteByID(id string) bool {
	_, ok := db.data[id]
	if !ok {
		return false
	}
	delete(db.data, id)
	return true
}
