package client

import (
	"encoding/json"
	piazza "github.com/venicegeo/pz-gocommon"
	"gopkg.in/olivere/elastic.v2"
	"log"
)

type ConditionDB struct {
	//data   map[string]Condition
	es    *piazza.ElasticSearchService
	index string
}

func NewConditionDB(es *piazza.ElasticSearchService, index string) (*ConditionDB, error) {
	db := new(ConditionDB)
	db.es = es
	db.index = index

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type XCondition struct {
	ID    string     `json:"id"`
	Title string    `json:"title"`
}

func (db *ConditionDB) Write(condition *Condition) error {

	id := NewConditionIdent()
	condition.ID = id

	xcondition := XCondition{ID: condition.ID.String(), Title: condition.Title}

	_, err := db.es.Client.Index().
		Index(db.index).
		Type("condition").
		Id(condition.ID.String()).
		BodyJson(xcondition).
		Do()
	if err != nil {
		log.Printf("baz 2 %#v", err)
		return err
	}

	err = db.es.Flush(db.index)
	if err != nil {
		log.Printf("baz 3 %#v", err)
		return err
	}

	log.Printf("baz 4")
	return nil
}

func (db *ConditionDB) Update(condition *Condition) bool {
	/**	_, ok := db.data[condition.ID]
	if ok {
		db.data[condition.ID] = *condition
		return true
	}
	**/
	return false
}

func (db *ConditionDB) ReadByID(id Ident) (*Condition, error) {
	get1, err := db.es.Client.Get().
	Index("conditions").
	Type("condition").
	Id("C1").
	Do()
	if err != nil {
		// Handle error
		panic(err)
	}
	log.Printf("!!!!!! %#v %#v", get1, get1.Source)
	b, err := get1.Source.MarshalJSON()
	if err != nil {
		panic("at disco")
	}
	log.Printf("bbbb %#v", b)


	termQuery := elastic.NewTermQuery("id", id.String())
	searchResult, err := db.es.Client.Search().
		Index(db.index).
		Query(termQuery).
		Do()

	if err != nil {
		return nil, err
	}
log.Printf("#1 %#v %s", id, db.index)
	for _, hit := range searchResult.Hits.Hits {
		log.Printf("#2")
		var a Condition
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			log.Printf("#3")
			return nil, err
		}
		log.Printf("#4 %#v", a)
		return &a, nil
	}

	log.Printf("#5")
	return nil, nil
}

func (db *ConditionDB) DeleteByID(id string) (bool, error) {
	res, err := db.es.Client.Delete().
		Index(db.index).
		Type("condition").
		Id(id).
		Do()
	if err != nil {
		return false, err
	}

	err = db.es.Flush(db.index)
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ConditionDB) GetAll() (map[Ident]Condition, error) {

	// search for everything
	// TODO: there's a GET call for this?
	searchResult, err := db.es.Client.Search().
		Index(db.index).
		Query(elastic.NewMatchAllQuery()).
		Sort("id", true).
		Do()
	if err != nil {
		return nil, err
	}

	m := make(map[Ident]Condition)

	for _, hit := range searchResult.Hits.Hits {
		var t Condition
		err := json.Unmarshal(*hit.Source, &t)
		if err != nil {
			return nil, err
		}
		m[t.ID] = t
	}

	return m, nil
}
