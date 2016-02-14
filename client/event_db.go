package client

import (
	"encoding/json"
)

var eventID = 1

func NewEventID() Ident {
	id := NewIdentFromInt(eventID)
	eventID++
	return Ident("E" + string(id))
}

//---------------------------------------------------------------------------


func ConvertRawsToEvents(raws []*json.RawMessage) ([]Event, error) {
	objs := make([]Event, len(raws))
	for i, _ := range raws {
		err := json.Unmarshal(*raws[i], &objs[i])
		if err != nil {
			return nil, err
		}
	}
	return objs, nil
}
