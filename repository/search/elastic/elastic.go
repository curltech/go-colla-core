package elastic

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/curltech/go-colla-core/config"
	baseentity "github.com/curltech/go-colla-core/entity"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/message"
	"github.com/curltech/go-colla-core/util/reflect"
	"github.com/dustin/go-humanize"
	elastic "github.com/elastic/go-elasticsearch/v8"
	esapi "github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type elasticSearchSession struct {
	es *elastic.Client
}

var ElasticSearchSession *elasticSearchSession = &elasticSearchSession{}

func (this *elasticSearchSession) Start() {
	if this.es != nil {
		return
	}
	cert, _ := ioutil.ReadFile(config.SearchParams.Cert)
	cfg := elastic.Config{
		Addresses: config.SearchParams.Address,
		Username:  config.SearchParams.Username,
		Password:  config.SearchParams.Password,
		CACert:    cert,
		//Transport: &fasthttp.Transport{},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   config.SearchParams.MaxIdleConns,
			ResponseHeaderTimeout: time.Duration(config.SearchParams.ResponseHeaderTimeout) * time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Nanosecond}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS11,
			},
		},
		// Retry on 429 TooManyRequests statuses
		//
		RetryOnStatus: []int{502, 503, 504, 429},

		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			retryBackoff := backoff.NewExponentialBackOff()
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 5 attempts
		//
		MaxRetries: config.SearchParams.MaxRetries,
	}
	es, err := elastic.NewClient(cfg)
	if err != nil {
		logger.Errorf("Error new elastic client: %s", err)
	}
	ElasticSearchSession.es = es
}

/**
显示客户端和服务端信息
*/
func (this *elasticSearchSession) Info() {
	res, err := this.es.Info()
	if err != nil {
		logger.Errorf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		logger.Errorf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	info := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		logger.Errorf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	logger.Infof("Client: %s", elastic.Version)
	logger.Infof("Server: %s", info["version"].(map[string]interface{})["number"])
}

/**
创建索引的批量创建器
*/
func (this *elasticSearchSession) newBulkIndexer(indexName string) esutil.BulkIndexer {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,                                                      // The default index name
		Client:        this.es,                                                        // The Elasticsearch client
		NumWorkers:    config.SearchParams.NumWorkers,                                 // The number of worker goroutines
		FlushBytes:    int(config.SearchParams.FlushBytes),                            // The flush threshold in bytes
		FlushInterval: time.Duration(config.SearchParams.FlushInterval) * time.Second, // The periodic flush interval
	})
	if err != nil {
		logger.Errorf("Error creating the indexer: %s", err)
	}

	return bi
}

func (this *elasticSearchSession) Index(indexName string, mds ...interface{}) error {
	var err error
	wg := sync.WaitGroup{}
	for _, md := range mds {
		if err != nil {
			return err
		}
		wg.Add(1)

		go func(md interface{}) {
			defer wg.Done()
			txt, err := message.TextMarshal(md)
			v, _ := reflect.GetValue(md, baseentity.FieldName_Id)
			id := fmt.Sprintf("%v", v)
			req := esapi.IndexRequest{
				Index:      indexName,
				DocumentID: id,
				Body:       strings.NewReader(txt),
				Refresh:    "true",
			}

			// Perform the request with the client.
			res, err := req.Do(context.Background(), this.es)
			if err != nil {
				logger.Errorf("Error getting response: %s", err)
				return
			}
			defer res.Body.Close()

			if res.IsError() {
				logger.Errorf("[%s] Error indexing document ID=%d", res.Status(), id)
				err = errors.New(fmt.Sprintf("StatusCode:%v", res.StatusCode))

				return

			} else {
				// Deserialize the response into a map.
				var r map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					logger.Infof("Error parsing the response body: %s", err)
				} else {
					// Print the response status and indexed document version.
					logger.Infof("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(md)
	}
	wg.Wait()

	return err
}

/**
标准检索json ES query查询
*/
func (this *elasticSearchSession) Query(indexName string, query string, from int, limit int) (map[string]interface{}, error) {
	//query := {
	//    "query": {
	//      "match": {
	//        "title": "test",
	//      },
	//    },
	//  }
	var buf bytes.Buffer
	_, err := buf.WriteString(query)
	if err != nil {
		logger.Errorf("Error encoding query: %s", err)

		return nil, err
	}

	// Perform the search request.
	res, err := this.es.Search(
		this.es.Search.WithContext(context.Background()),
		this.es.Search.WithIndex(indexName),
		this.es.Search.WithBody(&buf),
		this.es.Search.WithTrackTotalHits(true),
		this.es.Search.WithPretty(),
		this.es.Search.WithFrom(from),
		this.es.Search.WithSize(limit),
	)
	if err != nil {
		logger.Errorf("Error getting response: %s", err)

		return nil, err
	}
	defer res.Body.Close()

	return this.response(res, err)
}

func (this *elasticSearchSession) response(res *esapi.Response, err error) (map[string]interface{}, error) {
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			logger.Errorf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			logger.Errorf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}

		return nil, err
	}
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logger.Errorf("Error parsing the response body: %s", err)

		return nil, err
	}
	// Print the response status, number of results, and request duration.
	logger.Infof(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(result["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		logger.Infof(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	return result, nil
}

func (this *elasticSearchSession) CreateIndex(indexName string) {
	res, err := this.es.Indices.Create(indexName)
	if err != nil {
		logger.Errorf("Cannot create index: %s", err)
	}
	if res.IsError() {
		logger.Errorf("Cannot create index: %s", res)
	}
	res.Body.Close()
}

func (this *elasticSearchSession) DeleteIndex(indexName string) {
	res, err := this.es.Indices.Delete([]string{indexName}, this.es.Indices.Delete.WithIgnoreUnavailable(true))
	if err != nil || res.IsError() {
		logger.Errorf("Cannot delete index: %s", err)
	}
	res.Body.Close()
}

func (this *elasticSearchSession) BulkIndex(indexName string, mds ...interface{}) {
	var countSuccessful uint64
	bi := this.newBulkIndexer(indexName)
	for _, md := range mds {
		// Prepare the data payload: encode article to JSON
		data, err := json.Marshal(md)
		if err != nil {
			logger.Errorf("Cannot encode article %d: %s", md, err)
		}
		// Add an item to the BulkIndexer
		id, _ := reflect.GetValue(md, baseentity.FieldName_Id)
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				// DocumentID is the (optional) document ID
				DocumentID: strconv.Itoa(id.(int)),

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			logger.Errorf("Unexpected error: %s", err)
		}
	}
	// Close the indexer
	if err := bi.Close(context.Background()); err != nil {
		logger.Errorf("Unexpected error: %s", err)
	}

	biStats := bi.Stats()
	// Report the results: number of indexed docs, number of errors, duration, indexing rate
	if biStats.NumFailed > 0 {
		logger.Errorf(
			"Indexed [%s] documents with [%s] errors",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
		)
	} else {
		logger.Infof(
			"Sucessfuly indexed [%s] documents",
			humanize.Comma(int64(biStats.NumFlushed)))
	}
}

func (this *elasticSearchSession) Delete(indexName string, ids ...string) error {
	var err error
	wg := sync.WaitGroup{}
	for _, id := range ids {
		if err != nil {
			return err
		}
		wg.Add(1)

		go func(md interface{}) {
			defer wg.Done()
			req := esapi.DeleteRequest{
				Index:      indexName,
				DocumentID: fmt.Sprintf("%v", id),
				Refresh:    "true",
			}

			// Perform the request with the client.
			res, err := req.Do(context.Background(), this.es)
			if err != nil {
				logger.Errorf("Error getting response: %s", err)
				return
			}
			defer res.Body.Close()

			if res.IsError() {
				logger.Errorf("[%s] Error indexing document ID=%d", res.Status(), id)
				err = errors.New(fmt.Sprintf("StatusCode:%v", res.StatusCode))

				return
			}
		}(id)
	}
	wg.Wait()

	return err
}

func (this *elasticSearchSession) Get(indexName string, id string) (map[string]interface{}, error) {
	req := esapi.GetRequest{
		Index:      indexName,
		DocumentID: fmt.Sprintf("%v", id),
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), this.es)
	if err != nil {
		logger.Errorf("Error getting response: %s", err)

		return nil, err
	}
	if err != nil {
		logger.Errorf("Error getting response: %s", err)

		return nil, err
	}
	defer res.Body.Close()

	return this.response(res, err)
}

func (this *elasticSearchSession) Update(indexName string, mds ...interface{}) error {
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
