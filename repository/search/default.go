package search

import (
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/repository/search/elastic"
)

type SearchSession interface {
	Start()
	Index(indexName string, mds ...interface{}) error
	Delete(indexName string, ids ...string) error
	Update(indexName string, mds ...interface{}) error
	Get(indexName string, id string) (map[string]interface{}, error)
	Query(indexName string, query string, from int, limit int) (map[string]interface{}, error)
}

func GetSearchSession() SearchSession {
	if config.SearchParams.Mode == "bleve" {
		//bleve.BleveSession.Start()
		//
		//return bleve.BleveSession
	} else if config.SearchParams.Mode == "elastic" {
		elastic.ElasticSearchSession.Start()

		return elastic.ElasticSearchSession
	} else if config.SearchParams.Mode == "default" {
		elastic.DefaultSearchSession.Start()

		return elastic.DefaultSearchSession
	}
	return nil
}

func init() {

}
