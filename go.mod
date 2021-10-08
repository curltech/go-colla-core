module github.com/curltech/go-colla-core

go 1.17

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/ProtonMail/gopenpgp/v2 v2.0.1
	github.com/allegro/bigcache v1.2.1
	github.com/blevesearch/bleve/v2 v2.0.1
	github.com/blevesearch/bleve_index_api v1.0.0
	github.com/boltdb/bolt v1.3.1
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-elasticsearch/v8 v8.0.0-20200929065430-35fd8bce1107
	github.com/emersion/go-imap v1.0.6
	github.com/emersion/go-message v0.14.1
	github.com/emersion/go-msgauth v0.6.4
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21
	github.com/emersion/go-smtp v0.14.0
	github.com/ethereum/go-ethereum v1.9.21
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/hashicorp/go-hclog v0.9.1
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/raft v1.2.0
	github.com/hashicorp/raft-boltdb v0.0.0-20171010151810-6e5ba93211ea
	github.com/json-iterator/go v1.1.10
	github.com/lib/pq v1.8.0
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/multiformats/go-multiaddr v0.3.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olivere/elastic/v7 v7.0.22
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rqlite/rqlite v4.6.0+incompatible
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/streadway/amqp v1.0.0
	github.com/valyala/fasthttp v1.24.0
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.9
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2
	//github.com/yanyiwu/gojieba v1.1.2
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200909210914-44a2922940c2 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	gorm.io/driver/postgres v1.0.5
	gorm.io/gorm v1.20.5
	xorm.io/xorm v1.0.5
)

replace golang.org/x/crypto => github.com/ProtonMail/crypto v0.0.0-20200416114516-1fa7f403fb9c
