package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/curltech/go-colla-core/config"
	baseentity "github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/collection"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"sync"
)

type defaultSearchSession struct {
	es *elastic.Client
}

var DefaultSearchSession *defaultSearchSession = &defaultSearchSession{}

//初始化
func (this *defaultSearchSession) Start() {
	if this.es != nil {
		return
	}
	errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	es, err := elastic.NewClient(elastic.SetErrorLog(errorlog), elastic.SetURL(config.SearchParams.Address...))
	if err != nil {
		logger.Errorf("Error new elastic client: %s", err)
	}
	DefaultSearchSession.es = es
}

func (this *defaultSearchSession) Info() {
	info, code, err := this.es.Ping(config.SearchParams.Address[0]).Do(context.Background())
	if err != nil {
		logger.Errorf("Error ping: %s", err)
	}
	logger.Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := this.es.ElasticsearchVersion(config.SearchParams.Address[0])
	if err != nil {
		logger.Errorf("Error version info: %s", err)
	}
	logger.Infof("Elasticsearch version %s\n", esversion)
}

func (this *defaultSearchSession) Index(indexName string, mds ...interface{}) error {
	var err error
	wg := sync.WaitGroup{}
	for _, md := range mds {
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(md interface{}) {
			defer wg.Done()
			v, _ := reflect.GetValue(md, baseentity.FieldName_Id)
			id := fmt.Sprintf("%v", v)
			put1, err := this.es.Index().
				Index(indexName).
				Id(id).
				BodyJson(md).
				Do(context.Background())
			if err != nil {
				logger.Errorf("%v", err)
				return
			}
			logger.Infof("Indexed tweet %s to index s%s, type %s\n", put1.Id, put1.Index, put1.Type)
		}(md)
	}
	wg.Wait()

	return err
}

//删除
func (this *defaultSearchSession) Delete(indexName string, ids ...string) error {
	var err error
	wg := sync.WaitGroup{}
	for _, id := range ids {
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(md interface{}) {
			defer wg.Done()
			res, err := this.es.Delete().Index(indexName).
				Id(id).
				Do(context.Background())
			if err != nil {
				logger.Errorf("%v", err)

				return
			}
			logger.Infof("delete result %s\n", res.Result)
		}(id)
	}
	wg.Wait()

	return err
}

//修改
func (this *defaultSearchSession) Update(indexName string, mds ...interface{}) error {
	var err error
	wg := sync.WaitGroup{}
	for _, md := range mds {
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(md interface{}) {
			defer wg.Done()
			v, _ := reflect.GetValue(md, baseentity.FieldName_Id)
			id := fmt.Sprintf("%v", v)
			res, err := this.es.Update().
				Index(indexName).
				Id(id).
				Doc(md).
				Do(context.Background())
			if err != nil {
				logger.Errorf("%v", err)
				return
			}
			logger.Infof("update age %s\n", res.Result)
		}(md)
	}
	wg.Wait()

	return err
}

//查找
func (this *defaultSearchSession) Get(indexName string, id string) (map[string]interface{}, error) {
	//通过id查找
	result, err := this.es.Get().Index(indexName).Id(id).Do(context.Background())
	if err != nil {
		return nil, err
	}
	if result.Found {
		logger.Infof("Got document %s in version %d from index %s, type %s\n", result.Id, result.Version, result.Index, result.Type)
	}

	return collection.StructToMap(result, nil), nil
}

/**
标准检索json ES query查询
*/
func (this *defaultSearchSession) Query(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	var res *elastic.SearchResult
	var err error

	//字段相等
	q := elastic.NewRawStringQuery(query)
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}

	return this.response(res)
}

func (this *defaultSearchSession) StringQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	//QueryStringQuery("last_name:Smith")
	q := elastic.NewQueryStringQuery(query)
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}

	return this.response(res)
}

func (this *defaultSearchSession) MatchQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	q := elastic.NewBoolQuery()
	for k, v := range query {
		q.Must(elastic.NewMatchQuery(k, v))
	}
	var res *elastic.SearchResult
	var err error
	//q.Filter(elastic.NewRangeQuery("age").Gt(30))
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) MatchPhraseQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := elastic.NewBoolQuery()
	for k, v := range query {
		q.Must(elastic.NewMatchPhraseQuery(k, v))
	}
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) FuzzyQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := elastic.NewBoolQuery()
	for k, v := range query {
		q.Must(elastic.NewFuzzyQuery(k, v))
	}
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) TermQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := elastic.NewBoolQuery()
	for k, v := range query {
		q.Must(elastic.NewTermQuery(k, v))
	}
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) WildcardQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := elastic.NewBoolQuery()
	for k, v := range query {
		q.Must(elastic.NewWildcardQuery(k, v.(string)))
	}
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) PrefixQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := elastic.NewBoolQuery()
	for k, v := range query {
		q.Must(elastic.NewPrefixQuery(k, v.(string)))
	}
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) RangeQuery(indexName string, query map[string]map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := elastic.NewBoolQuery()
	for k, r := range query {
		v, ok := r["min"]
		if ok {
			min := v.(float64)
			v, ok = r["max"]
			if ok {
				max := v.(float64)
				q.Must(elastic.NewRangeQuery(k).From(min).To(max))
			}
		}
	}
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Query(q).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) TermsAggregation(indexName string, field string, value string, from int, limit int) (map[string]interface{}, error) {
	//分析 interests
	aggs := elastic.NewTermsAggregation().Field(field)
	var res *elastic.SearchResult
	var err error
	res, err = this.es.Search(indexName).Size(limit).From(from).Aggregation(value, aggs).Do(context.Background())
	if err != nil {
		logger.Errorf("%v", err)
	}
	return this.response(res)
}

func (this *defaultSearchSession) response(res *elastic.SearchResult) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	//for _, item := range res.Each(basereflect.TypeOf(result)) { //从搜索结果中取数据的方法
	//	result = item.(map[string]interface{})
	//}
	if res.Hits.TotalHits.Value > 0 {
		logger.Infof("Found a total of %d \n", res.Hits.TotalHits)
		for _, hit := range res.Hits.Hits {
			buf, err := hit.Source.MarshalJSON()
			if err != nil {
				logger.Infof("Deserialization failed")
			}
			err = json.Unmarshal(buf, &result) //另外一种取数据的方法
			if err != nil {
				logger.Infof("Deserialization failed")
			}
		}
	} else {
		logger.Infof("Found no result \n")
	}

	return result, nil
}
