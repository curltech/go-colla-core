package bleve

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	blevequery "github.com/blevesearch/bleve/search/query"
	"github.com/curltech/go-colla-core/config"
	baseentity "github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/collection"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/ttys3/gojieba-bleve"
	"github.com/yanyiwu/gojieba"
	"os"
)

type bleveSession struct {
	indexMapping mapping.IndexMapping
	indexes      map[string]bleve.Index
}

var BleveSession = &bleveSession{}

func (this *bleveSession) Start() {
	if this.indexes == nil {
		this.indexes = make(map[string]bleve.Index)
	}
	indexMapping := bleve.NewIndexMapping()
	this.zh(indexMapping)
	this.indexMapping = indexMapping
}

func (this *bleveSession) init(indexName string) error {
	_, ok := this.indexes[indexName]
	if ok {
		return nil
	}
	var err error
	var index bleve.Index
	var path = config.SearchParams.Address[0] + "." + indexName
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//index, err = bleve.New(path, indexMapping)
		index, err = bleve.NewUsing(path, this.indexMapping, scorch.Name, scorch.Name, nil)
		if err != nil {
			return err
		}
	} else {
		index, err = bleve.OpenUsing(path, map[string]interface{}{
			"create_if_missing": false,
			"error_if_exists":   false,
		})
	}
	if err != nil {
		return err
	}
	BleveSession.indexes[indexName] = index

	return nil
}

func (this *bleveSession) zh(indexMapping *mapping.IndexMappingImpl) error {
	err := indexMapping.AddCustomTokenizer("gojieba",
		map[string]interface{}{
			"dictpath":     gojieba.DICT_PATH,
			"hmmpath":      gojieba.HMM_PATH,
			"userdictpath": gojieba.USER_DICT_PATH,
			"idf":          gojieba.IDF_PATH,
			"stop_words":   gojieba.STOP_WORDS_PATH,
			"type":         "gojieba",
		},
	)
	if err != nil {
		return err
	}
	err = indexMapping.AddCustomAnalyzer("gojieba",
		map[string]interface{}{
			"type":      "gojieba",
			"tokenizer": "gojieba",
		},
	)
	if err != nil {
		return err
	}
	indexMapping.DefaultAnalyzer = "gojieba"

	return nil
}

func (this *bleveSession) AddDocumentMapping(doctype string, dm *mapping.DocumentMapping) {
	//blogMapping := bleve.NewDocumentMapping()
	//this.engine.Mapping().AddDocumentMapping("blog", blogMapping)
}

// Get retrieve one record from database, bean's non-empty fields
// will be as conditions
func (this *bleveSession) Get(indexName string, id string) (map[string]interface{}, error) {
	err := this.init(indexName)
	if err != nil {
		return nil, err
	}
	doc, err := this.indexes[indexName].Document(id)
	if err == nil && doc != nil {
		result := make(map[string]interface{})
		for _, field := range doc.Fields {
			result[field.Name()] = (string)(field.Value())
		}
		logger.Infof("%v", result)

		return result, nil
	}

	return nil, err
}

// Find retrieve records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct everyone := make([]Userinfo, 0)
// err := engine.Find(&everyone)
func (this *bleveSession) Query(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	var q blevequery.Query = bleve.NewMatchQuery(query)

	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) query(indexName string, q blevequery.Query, from int, limit int) (map[string]interface{}, error) {
	err := this.init(indexName)
	if err != nil {
		return nil, err
	}
	req := bleve.NewSearchRequest(q)
	req.Highlight = bleve.NewHighlight()
	req.From = from
	if limit > 0 {
		req.Size = limit
	}
	searchResults, err := this.indexes[indexName].Search(req)
	if err != nil {
		return nil, err
	}
	logger.Infof("%v", searchResults)
	result := this.response(searchResults)

	return result, nil
}

// insert model data to database
func (this *bleveSession) Index(indexName string, mds ...interface{}) error {
	err := this.init(indexName)
	if err != nil {
		return err
	}
	batch := this.indexes[indexName].NewBatch()
	for _, md := range mds {
		v, _ := reflect.GetValue(md, baseentity.FieldName_Id)
		id := fmt.Sprintf("%v", v)
		err := batch.Index(fmt.Sprintf("%v", id), md)
		if err != nil {
			logger.Errorf("%v", err)
		}
	}
	err = this.indexes[indexName].Batch(batch)

	return err
}

// delete model in database
// Delete records, bean's non-empty fields are conditions
func (this *bleveSession) Delete(indexName string, ids ...string) error {
	err := this.init(indexName)
	if err != nil {
		return err
	}
	batch := this.indexes[indexName].NewBatch()
	for _, id := range ids {
		batch.Delete(fmt.Sprintf("%v", id))
	}
	err = this.indexes[indexName].Batch(batch)

	return err
}

func (this *bleveSession) Update(indexName string, mds ...interface{}) error {
	ids := make([]string, 0)
	for _, md := range mds {
		v, _ := reflect.GetValue(md, baseentity.FieldName_Id)
		id := fmt.Sprintf("%v", v)
		ids = append(ids, id)
	}
	err := this.Delete(indexName, ids...)
	if err == nil {
		err = this.Index(indexName, mds...)
	}

	return err
}

func (this *bleveSession) Close(indexName string) error {
	err := this.init(indexName)
	if err != nil {
		return err
	}
	analyzer := this.indexes[indexName].Mapping().AnalyzerNamed("gojieba")
	tokenizer, ok := analyzer.Tokenizer.(*jbleve.JiebaTokenizer)
	if !ok {
		panic("jieba.Free() failed")
	} else {
		tokenizer.Free()
	}
	return this.indexes[indexName].Close()
}

func (this *bleveSession) response(searchResult *bleve.SearchResult) map[string]interface{} {
	result := collection.StructToMap(searchResult, nil)

	return result
}

func (this *bleveSession) MatchPhraseQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	q := bleve.NewMatchPhraseQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) QueryStringQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	q := bleve.NewQueryStringQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) MatchQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	q := bleve.NewMatchQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) FuzzyQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := bleve.NewFuzzyQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) TermQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := bleve.NewTermQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) WildcardQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := bleve.NewWildcardQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) PrefixQuery(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	//短语搜索 搜索about字段中有 rock climbing
	q := bleve.NewPrefixQuery(query)
	return this.query(indexName, q, from, limit)
}

func (this *bleveSession) RangeQuery(indexName string, query map[string]interface{}, from int, limit int) (map[string]interface{}, error) {
	v, ok := query["min"]
	if ok {
		min := v.(float64)
		v, ok = query["max"]
		if ok {
			max := v.(float64)
			q := bleve.NewNumericRangeQuery(&min, &max)
			return this.query(indexName, q, from, limit)
		}
	}

	return nil, errors.New("ErrorQuery")
}
