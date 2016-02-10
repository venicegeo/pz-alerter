package client

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v2"
	"github.com/venicegeo/pz-gocommon"
	"log"
)

//---------------------------------------------------------------------------

type ActionDB struct {
	es *piazza.ElasticSearchService
	index  string
}

func NewActionDB(es *piazza.ElasticSearchService, index string) (*ActionDB, error) {
	db := new(ActionDB)
	db.es = es
	db.index = index

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *ActionDB) Write(action *Action) error {

	_, err := db.es.Client.Index().
		Index(db.index).
		Type("action").
		Id(action.ID.String()).
		BodyJson(action).
		Do()
	if err != nil {
		return err
	}

	err = db.es.Flush(db.index)
	if err != nil {
		return err
	}

	return nil
}


func (db *ActionDB) GetAll() (map[Ident]Action, error) {
	searchResult, err := db.es.Client.Search().
		Index(db.index).
		Query(elastic.NewMatchAllQuery()).
		Sort("id", true).
		Do()
	if err != nil {
		return nil, err
	}

	m := make(map[Ident]Action)

	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var t Action
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				return nil, err
			}
			m[t.ID] = t
		}
	}

	return m, nil
}

func (db *ActionDB) GetByID(id Ident) (*Action, error) {

	m, err := db.GetAll()
	if err != nil {
		log.Fatalf("bonk")
	}
	log.Printf("ALL: %#v", m)



	err = db.es.Flush(db.index)
	if err != nil {
		log.Print("done -1")
		return nil, err
	}


	termQuery := elastic.NewTermQuery("id", id.String())
	//termQuery := elastic.NewMatchAllQuery()
	searchResult, err := db.es.Client.Search().
	Index(db.index).
	Query(termQuery).
	Sort("id", false).
	Do()

	if err != nil {
		log.Print("done 0", err)
		return nil, err
	}
	log.Print("**target ", id.String())

	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			log.Print("****target ", id)
			var a Action
			err := json.Unmarshal(*hit.Source, &a)
			log.Printf("**hit %#v", hit)
			if err != nil {
				return nil, err
			}
			log.Print("done 1")
			//return &a, nil
		}
	}

	log.Print("done 2")
	return nil, nil
}
