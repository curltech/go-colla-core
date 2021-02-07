package config

import (
	"strings"
	"time"
)

type appParams struct {
	Enable        bool
	P2pProtocol   string
	TimeFormat    string
	EnableSession bool
	SessionLog    bool
	Template      string
	Name          string
}

type p2pParams struct {
	ChainProtocolID string
}

type consensusParams struct {
	PeerRange      int
	PeerNum        int
	StdMinPeerNum  int
	RaftMinPeerNum int
	Selector       string
}

type databaseParams struct {
	Drivername      string
	Host            string
	Port            string
	Dbname          string
	User            string
	Password        string
	Sslmode         string
	TimeZone        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	ShowSQL         bool
	LogLevel        int
	Readtransaction bool
	Orm             string
	Dsn             string
	Sequence        string //sequence的产生方式，缺省是seq，可选table
}

type searchParams struct {
	Address               []string
	Username              string
	Password              string
	MaxIdleConns          int
	ResponseHeaderTimeout int
	MaxRetries            int
	Cert                  string
	NumWorkers            int
	FlushBytes            int
	FlushInterval         int
	Mode                  string
}

type rqliteParams struct {
	HttpAddr               string
	HttpAdv                string
	AuthFile               string
	X509CACert             string
	X509Cert               string
	X509Key                string
	NodeEncrypt            bool
	NodeX509CACert         string
	NodeX509Cert           string
	NodeX509Key            string
	NodeID                 string
	RaftAddr               string
	RaftAdv                string
	JoinAddr               string
	JoinAttempts           int
	JoinInterval           string
	NoVerify               bool
	NoNodeVerify           bool
	DiscoURL               string
	DiscoID                string
	Expvar                 bool
	PprofEnabled           bool
	Dsn                    string
	OnDisk                 bool
	RaftLogLevel           string
	RaftNonVoter           bool
	RaftSnapThreshold      uint64
	RaftSnapInterval       string
	RaftLeaderLeaseTimeout string
	RaftHeartbeatTimeout   string
	RaftElectionTimeout    string
	RaftApplyTimeout       string
	RaftOpenTimeout        string
	RaftShutdownOnRemove   bool
	CompressionSize        int
	CompressionBatch       int
	ShowVersion            bool
	CpuProfile             string
	MemProfile             string
}

type libp2pParams struct {
	Enable                   bool
	Topic                    string
	Addrs                    []string
	Addr                     string
	Port                     string
	WsPort                   string
	ReadTimeout              int
	WriteTimeout             int
	EnableTls                bool
	EnableSecio              bool
	EnableNoise              bool
	EnableQuic               bool
	LowWater                 int
	HighWater                int
	GracePeriod              int
	EnableNatPortMap         bool
	EnableRouting            bool
	EnableRelay              bool
	EnableAutoRelay          bool
	EnableWebsocket          bool
	EnableWebrtc             bool
	EnableAutoNat            bool
	EnableNATService         bool
	ConnectionGater          bool
	ForceReachabilityPublic  bool
	ForceReachabilityPrivate bool
	EnableAddressFactory     bool
	ExternalAddr             string
	ExternalPort             string
	ExternalWsPort           string
	ExternalWssPort          string
	FaultTolerantLevel       int
	Nvals                    int
	Quorum                   int
}

/**
专用于iris的配置
*/
type serverParams struct {
	Addr         string
	Port         string
	Password     string
	Name         string
	Email        string
	ExternalAddr string
	ExternalPort string
}

type serverWebsocketParams struct {
	Mode             string
	Address          string
	ReadBufferSize   int
	WriteBufferSize  int
	Path             string
	HeartbeatInteval int64
}

type ipfsParams struct {
	Enable              bool
	RepoPath            string
	ExternalPluginsPath string
	BootstrapNodes      []string
}

type tlsParams struct {
	Mode   string
	Port   string
	Url    string
	Email  string
	Cert   string
	Key    string
	Domain string
}

type proxyParams struct {
	Mode     string
	Address  string
	Target   string
	Redirect bool
}

type rbacParams struct {
	NonePath    []string
	NoneAddress []string
	Model       string
	/**
	只有在resource表中存在的资源才校验权限，否则，不校验
	*/
	ValidResource      bool
	Credential         string
	Password           string
	AccessTokenMaxAge  int64
	RefreshLeftAge     int64
	RefreshTokenMaxAge int64
	PrivateKeyFileName string
	PublicKeyFileName  string
}

type turnParams struct {
	Enable      bool
	Ip          string
	Host        string
	TcpPort     string
	UdpPort     string
	Realm       string
	Credentials string
	Cert        string
	Key         string
}

type sfuParams struct {
	Enable              bool
	Ballast             int64
	Withstats           bool
	Maxbandwidth        uint64
	Maxbuffertime       int
	Bestqualityfirst    bool
	Enabletemporallayer bool
	Minport             uint16
	Maxport             uint16
	Sdpsemantics        string
	Level               string
	Urls                [][]string
	Usernames           []string
	Credentials         []string
}

type smtpServerParams struct {
	Enable          bool
	Addr            string
	Domain          string
	ReadTimeout     int64
	WriteTimeout    int64
	MaxMessageBytes uint64
	MaxRecipients   int
}

type imapServerParams struct {
	Enable bool
	Addr   string
	Domain string
}

var AppParams = appParams{}

var ServerWebsocketParams = serverWebsocketParams{}

var DatabaseParams = databaseParams{}

var SearchParams = searchParams{}

var RqliteParams = rqliteParams{}

var P2pParams = p2pParams{}

var ConsensusParams = consensusParams{}

var Libp2pParams = libp2pParams{}

var ServerParams = serverParams{}

var IpfsParams = ipfsParams{}

var TlsParams = tlsParams{}

var ProxyParams = proxyParams{}

var RbacParams = rbacParams{}

var TurnParams = turnParams{}

var SfuParams = sfuParams{}

var SmtpServerParams = smtpServerParams{}

var ImapServerParams = imapServerParams{}

func init() {
	AppParams.Enable, _ = GetBool("app.enable", true)
	AppParams.P2pProtocol, _ = GetString("app.p2pProtocol", "libp2p")
	AppParams.TimeFormat, _ = GetString("app.TimeFormat", time.RFC3339Nano)
	AppParams.EnableSession, _ = GetBool("app.enableSession", false)
	if AppParams.EnableSession {
		AppParams.SessionLog, _ = GetBool("app.sessionLog", false)
	} else {
		AppParams.SessionLog = false
	}
	AppParams.Template, _ = GetString("app.template", "html")

	P2pParams.ChainProtocolID, _ = GetString("p2p.chainProtocolID", "/chain/1.0.0")

	ConsensusParams.PeerRange, _ = GetInt("consensus.peerRange", 10)
	ConsensusParams.PeerNum, _ = GetInt("consensus.peerNum", 4)
	ConsensusParams.StdMinPeerNum, _ = GetInt("consensus.stdMinPeerNum", 0)
	ConsensusParams.RaftMinPeerNum, _ = GetInt("consensus.raftMinPeerNum", 1)
	ConsensusParams.Selector, _ = GetString("consensus.selector", "random")

	Libp2pParams.Enable, _ = GetBool("libp2p.enable", false)
	addrs, _ := GetString("libp2p.addrs", "")
	if addrs != "" {
		Libp2pParams.Addrs = strings.Split(addrs, ",")
	}
	Libp2pParams.Addr, _ = GetString("libp2p.addr", "0.0.0.0")
	Libp2pParams.Port, _ = GetString("libp2p.port", "3719")

	Libp2pParams.Topic, _ = GetString("libp2p.topic", "")
	Libp2pParams.ReadTimeout, _ = GetInt("libp2p.readTimeout", 5000)
	Libp2pParams.WriteTimeout, _ = GetInt("libp2p.writeTimeout", 5000)
	Libp2pParams.EnableTls, _ = GetBool("libp2p.enableTls", true)
	Libp2pParams.EnableSecio, _ = GetBool("libp2p.enableSecio", false)
	Libp2pParams.EnableNoise, _ = GetBool("libp2p.enableNoise", true)
	Libp2pParams.EnableQuic, _ = GetBool("libp2p.enableQuic", false)
	Libp2pParams.EnableNatPortMap, _ = GetBool("libp2p.enableNatPortMap", true)
	Libp2pParams.EnableRouting, _ = GetBool("libp2p.enableRouting", true)
	Libp2pParams.EnableRelay, _ = GetBool("libp2p.enableRelay", true)
	Libp2pParams.EnableAutoRelay, _ = GetBool("libp2p.enableAutoRelay", false)
	//这个参数表示只启动websocket，不启动Tcp，缺省Tcp，websocket是启动的
	Libp2pParams.EnableWebsocket, _ = GetBool("libp2p.enableWebsocket", false)
	Libp2pParams.WsPort, _ = GetString("libp2p.wsPort", "4719")
	Libp2pParams.EnableWebrtc, _ = GetBool("libp2p.enableWebrtcStar", false)
	Libp2pParams.EnableWebrtc, _ = GetBool("libp2p.enableWebrtc", false)
	Libp2pParams.EnableAutoNat, _ = GetBool("libp2p.enableAutoNat", false)
	Libp2pParams.EnableNATService, _ = GetBool("libp2p.enableNATService", false)
	Libp2pParams.ConnectionGater, _ = GetBool("libp2p.connectionGater", false)
	Libp2pParams.ForceReachabilityPublic, _ = GetBool("libp2p.forceReachabilityPublic", false)
	Libp2pParams.ForceReachabilityPrivate, _ = GetBool("libp2p.forceReachabilityPrivate", false)

	Libp2pParams.LowWater, _ = GetInt("libp2p.LowWater", 100)
	Libp2pParams.HighWater, _ = GetInt("libp2p.HighWater", 400)
	Libp2pParams.GracePeriod, _ = GetInt("libp2p.GracePeriod", 1)
	Libp2pParams.EnableAddressFactory, _ = GetBool("libp2p.enableAddressFactory", true)
	Libp2pParams.ExternalAddr, _ = GetString("libp2p.externalAddr")
	Libp2pParams.ExternalPort, _ = GetString("libp2p.externalPort", "3720")
	Libp2pParams.ExternalWsPort, _ = GetString("libp2p.externalWsPort", "4720")
	Libp2pParams.ExternalWssPort, _ = GetString("libp2p.externalWssPort", "5720")

	Libp2pParams.FaultTolerantLevel, _ = GetInt("libp2p.faultTolerantLevel", 0)
	Libp2pParams.Nvals, _ = GetInt("libp2p.nvals", 1)
	Libp2pParams.Quorum, _ = GetInt("libp2p.quorum", 0)

	ServerParams.Addr, _ = GetString("http.addr")
	ServerParams.Port, _ = GetString("http.port", "8080")
	ServerParams.ExternalAddr, _ = GetString("http.externalAddr", "0.0.0.0")
	ServerParams.ExternalPort, _ = GetString("http.externalPort", "8089")
	ServerParams.Name, _ = GetString("server.name")
	ServerParams.Password, _ = GetString("server.password")
	ServerParams.Email, _ = GetString("server.email")

	DatabaseParams.Drivername, _ = GetString("database.drivername", "postgres")
	DatabaseParams.Dsn, _ = GetString("database.dsn")
	DatabaseParams.Dbname, _ = GetString("database.dbname", "postgres")
	DatabaseParams.Host, _ = GetString("database.host", "localhost")
	DatabaseParams.Port, _ = GetString("database.port", "5432")
	DatabaseParams.User, _ = GetString("database.user", "postgres")
	DatabaseParams.Password, _ = GetString("database.password")
	DatabaseParams.Readtransaction, _ = GetBool("database.readtransaction", false)
	DatabaseParams.Sslmode, _ = GetString("database.sslmode")
	DatabaseParams.TimeZone, _ = GetString("database.timeZone")
	DatabaseParams.MaxIdleConns, _ = GetInt("database.maxIdleConns")
	DatabaseParams.MaxOpenConns, _ = GetInt("database.maxOpenConns")
	DatabaseParams.ConnMaxLifetime, _ = GetInt("database.connMaxLifetime")
	DatabaseParams.ShowSQL, _ = GetBool("database.showSQL", false)
	DatabaseParams.Orm, _ = GetString("database.orm", "xorm")
	DatabaseParams.Sequence, _ = GetString("database.sequence", "seq")
	level, _ := GetString("database.logLevel", "info")
	switch level {
	case "debug":
		DatabaseParams.LogLevel = 0
	case "info":
		DatabaseParams.LogLevel = 1
	case "warn":
		DatabaseParams.LogLevel = 2
	case "error":
		DatabaseParams.LogLevel = 3
	case "off":
		DatabaseParams.LogLevel = 4
	}

	SearchParams.Mode, _ = GetString("search.mode", "bleve")
	if SearchParams.Mode == "default" || SearchParams.Mode == "elastic" {
		address, _ := GetString("search.address", "http://localhost:9200")
		if address != "" {
			SearchParams.Address = strings.Split(address, " ")
		}
	} else if SearchParams.Mode == "bleve" {
		address, _ := GetString("search.address", "bleve")
		if address != "" {
			SearchParams.Address = strings.Split(address, " ")
		}
	}
	SearchParams.Username, _ = GetString("search.username")
	SearchParams.Password, _ = GetString("search.password")
	SearchParams.Cert, _ = GetString("search.cert")
	SearchParams.MaxIdleConns, _ = GetInt("search.maxIdleConns", 10)
	SearchParams.ResponseHeaderTimeout, _ = GetInt("search.responseHeaderTimeout", 30)
	SearchParams.MaxRetries, _ = GetInt("search.maxRetries", 5)
	SearchParams.NumWorkers, _ = GetInt("search.indexer.numWorkers", 5)
	SearchParams.FlushBytes, _ = GetInt("search.indexer.flushBytes", 1024*8)
	SearchParams.FlushInterval, _ = GetInt("search.indexer.flushInterval", 30)

	RqliteParams.NodeID, _ = GetString("node-id", "", "Unique name for node. If not set, set to hostname")
	RqliteParams.HttpAddr, _ = GetString("http-addr", "localhost:4001", "HTTP server bind address. For HTTPS, set X.509 cert and key")
	RqliteParams.HttpAdv, _ = GetString("http-adv-addr", "", "Advertised HTTP address. If not set, same as HTTP server")
	RqliteParams.X509CACert, _ = GetString("http-ca-cert", "", "Path to root X.509 certificate for HTTP endpoint")
	RqliteParams.X509Cert, _ = GetString("http-cert", "", "Path to X.509 certificate for HTTP endpoint")
	RqliteParams.X509Key, _ = GetString("http-key", "", "Path to X.509 private key for HTTP endpoint")
	RqliteParams.NoVerify, _ = GetBool("http-no-verify", false)
	RqliteParams.NodeEncrypt, _ = GetBool("node-encrypt", false)
	RqliteParams.NodeX509CACert, _ = GetString("node-ca-cert", "", "Path to root X.509 certificate for node-to-node encryption")
	RqliteParams.NodeX509Cert, _ = GetString("node-cert", "cert.pem", "Path to X.509 certificate for node-to-node encryption")
	RqliteParams.NodeX509Key, _ = GetString("node-key", "key.pem", "Path to X.509 private key for node-to-node encryption")
	RqliteParams.NoNodeVerify, _ = GetBool("node-no-verify", false)
	RqliteParams.AuthFile, _ = GetString("auth", "", "Path to authentication and authorization file. If not set, not enabled")
	RqliteParams.RaftAddr, _ = GetString("raft-addr", "localhost:4002", "Raft communication bind address")
	RqliteParams.RaftAdv, _ = GetString("raft-adv-addr", "", "Advertised Raft communication address. If not set, same as Raft bind")
	RqliteParams.JoinAddr, _ = GetString("join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	RqliteParams.JoinAttempts, _ = GetInt("join-attempts", 5)
	RqliteParams.JoinInterval, _ = GetString("join-interval", "5s", "Period between join attempts")
	RqliteParams.DiscoURL, _ = GetString("disco-url", "http://discovery.rqlite.com", "Set Discovery Service URL")
	RqliteParams.DiscoID, _ = GetString("disco-id", "", "Set Discovery ID. If not set, Discovery Service not used")
	RqliteParams.Expvar, _ = GetBool("expvar", true)
	RqliteParams.PprofEnabled, _ = GetBool("pprof", true)
	RqliteParams.Dsn, _ = GetString("dsn", "", `SQLite DSN parameters. E.g. "cache=shared&mode=memory"`)
	RqliteParams.OnDisk, _ = GetBool("on-disk", false)
	RqliteParams.ShowVersion, _ = GetBool("version", false)
	RqliteParams.RaftNonVoter, _ = GetBool("raft-non-voter", false)
	RqliteParams.RaftHeartbeatTimeout, _ = GetString("raft-timeout", "1s", "Raft heartbeat timeout")
	RqliteParams.RaftElectionTimeout, _ = GetString("raft-election-timeout", "1s", "Raft election timeout")
	RqliteParams.RaftApplyTimeout, _ = GetString("raft-apply-timeout", "10s", "Raft apply timeout")
	RqliteParams.RaftOpenTimeout, _ = GetString("raft-open-timeout", "120s", "Time for initial Raft logs to be applied. Use 0s duration to skip wait")
	RqliteParams.RaftSnapThreshold, _ = GetUint64("raft-snap", 8192)
	RqliteParams.RaftSnapInterval, _ = GetString("raft-snap-int", "30s", "Snapshot threshold check interval")
	RqliteParams.RaftLeaderLeaseTimeout, _ = GetString("raft-leader-lease-timeout", "0s", "Raft leader lease timeout. Use 0s for Raft default")
	RqliteParams.RaftShutdownOnRemove, _ = GetBool("raft-remove-shutdown", false)
	RqliteParams.RaftLogLevel, _ = GetString("raft-log-level", "INFO", "Minimum log level for Raft module")
	RqliteParams.CompressionSize, _ = GetInt("compression-size", 150)
	RqliteParams.CompressionBatch, _ = GetInt("compression-batch", 5)
	RqliteParams.CpuProfile, _ = GetString("cpu-profile", "", "Path to file for CPU profiling information")
	RqliteParams.MemProfile, _ = GetString("mem-profile", "", "Path to file for memory profiling information")

	ServerWebsocketParams.Mode, _ = GetString("server.websocket.mode", "iris")
	ServerWebsocketParams.Address, _ = GetString("server.websocket.address", ":9090")
	ServerWebsocketParams.ReadBufferSize, _ = GetInt("server.websocket.readBufferSize", 4096)
	ServerWebsocketParams.WriteBufferSize, _ = GetInt("server.websocket.writeBufferSize", 1024)
	ServerWebsocketParams.Path, _ = GetString("server.websocket.path", "/websocket")
	ServerWebsocketParams.HeartbeatInteval, _ = GetInt64("server.websocket.heartbeatInteval", 2)

	IpfsParams.Enable, _ = GetBool("ipfs.enable", false)
	IpfsParams.RepoPath, _ = GetString("ipfs.repoPath", "")
	IpfsParams.ExternalPluginsPath, _ = GetString("ipfs.repoPath")
	bootstrapNode, _ := GetString("ipfs.repoPath")
	if bootstrapNode != "" {
		IpfsParams.BootstrapNodes = strings.Split(bootstrapNode, ",")
	}

	TlsParams.Mode, _ = GetString("http.tls.mode", "none")
	TlsParams.Port, _ = GetString("http.tls.port", "9090")
	TlsParams.Cert, _ = GetString("http.tls.cert", "conf/camsi-server-ec.crt")
	TlsParams.Key, _ = GetString("http.tls.key", "conf/camsi-server-ec.key")
	TlsParams.Url, _ = GetString("http.tls.url")
	TlsParams.Email, _ = GetString("http.tls.email")
	TlsParams.Domain, _ = GetString("http.tls.domain")

	ProxyParams.Mode, _ = GetString("http.proxy.mode", "none")
	ProxyParams.Address, _ = GetString("http.proxy.address", ":9090")
	ProxyParams.Target, _ = GetString("http.proxy.target", "none")
	ProxyParams.Redirect, _ = GetBool("http.proxy.redirect")

	nonePath, _ := GetString("rbac.nonePath", "/user/Login,/user/Logout")
	if nonePath != "" {
		RbacParams.NonePath = strings.Split(nonePath, ",")
	}
	noneAddress, _ := GetString("rbac.noneAddress", "127.0.0.1")
	if noneAddress != "" {
		RbacParams.NoneAddress = strings.Split(noneAddress, ",")
	}
	RbacParams.Model, _ = GetString("rbac.model", "conf/rbac_model.conf")
	RbacParams.ValidResource, _ = GetBool("rbac.validResource", true)
	RbacParams.Credential, _ = GetString("rbac.userName", "credential_")
	RbacParams.Password, _ = GetString("rbac.password", "password_")
	accessTokenMaxAge, _ := GetInt64("rbac.accessTokenMaxAge", 10)
	RbacParams.AccessTokenMaxAge = accessTokenMaxAge * int64(time.Minute)
	refreshLeftAge, _ := GetInt64("rbac.refreshLeftAge", 5)
	RbacParams.RefreshLeftAge = refreshLeftAge * int64(time.Minute)
	refreshTokenMaxAge, _ := GetInt64("rbac.refreshTokenMaxAge", 10)
	RbacParams.RefreshTokenMaxAge = refreshTokenMaxAge * int64(time.Hour)
	RbacParams.PrivateKeyFileName, _ = GetString("rbac.privateKeyFileName", "conf/ed25519_private_key.pem")
	RbacParams.PublicKeyFileName, _ = GetString("rbac.publicKeyFileName", "conf/ed25519_public_key.pem")

	TurnParams.Enable, _ = GetBool("turn.enable", false)
	TurnParams.Host, _ = GetString("turn.host", "localhost")
	TurnParams.TcpPort, _ = GetString("turn.tcpPort", "3478")
	TurnParams.UdpPort, _ = GetString("turn.udpPort", "3478")
	TurnParams.Realm, _ = GetString("turn.realm", "pion.ly")
	TurnParams.Ip, _ = GetString("turn.ip", "127.0.0.1")
	TurnParams.Credentials, _ = GetString("turn.credentials")
	TurnParams.Cert, _ = GetString("turn.cert")
	TurnParams.Key, _ = GetString("turn.key")

	SfuParams.Enable, _ = GetBool("sfu.enable", false)
	SfuParams.Ballast, _ = GetInt64("sfu.ballast", 0)
	SfuParams.Withstats, _ = GetBool("sfu.withstats", false)
	SfuParams.Maxbandwidth, _ = GetUint64("sfu.maxbandwidth")
	SfuParams.Maxbuffertime, _ = GetInt("sfu.maxbuffertime")
	SfuParams.Bestqualityfirst, _ = GetBool("sfu.bestqualityfirst", true)
	SfuParams.Enabletemporallayer, _ = GetBool("sfu.enabletemporallayer")
	SfuParams.Minport, _ = GetUint16("sfu.minport")
	SfuParams.Maxport, _ = GetUint16("sfu.maxport")
	SfuParams.Sdpsemantics, _ = GetString("sfu.sdpsemantics")
	SfuParams.Level, _ = GetString("sfu.level")
	SfuParams.Urls = make([][]string, 0)
	SfuParams.Usernames = make([]string, 0)
	SfuParams.Credentials = make([]string, 0)
	urls, _ := GetString("sfu.urls")
	if urls != "" {
		url := strings.Split(urls, ":")
		usernames := strings.Split(urls, ":")
		credentials := strings.Split(urls, ":")
		for i, u := range url {
			SfuParams.Urls = append(SfuParams.Urls, strings.Split(u, ","))
			if i < len(usernames) {
				SfuParams.Usernames = append(SfuParams.Usernames, usernames[i])
			}
			if i < len(credentials) {
				SfuParams.Credentials = append(SfuParams.Credentials, credentials[i])
			}
		}
	}

	SmtpServerParams.Enable, _ = GetBool("mail.server.smtp.enable", false)
	SmtpServerParams.Addr, _ = GetString("mail.server.smtp.addr", ":1025")
	SmtpServerParams.Domain, _ = GetString("mail.server.smtp.domain", "localhost")
	SmtpServerParams.ReadTimeout, _ = GetInt64("mail.server.smtp.readTimeout", 10*int64(time.Second))
	SmtpServerParams.WriteTimeout, _ = GetInt64("mail.server.smtp.writeTimeout", 10*int64(time.Second))
	SmtpServerParams.MaxMessageBytes, _ = GetUint64("mail.server.smtp.maxMessageBytes", 1024*1024)
	SmtpServerParams.MaxRecipients, _ = GetInt("mail.server.smtp.maxRecipients", 50)

	ImapServerParams.Enable, _ = GetBool("mail.server.imap.enable", false)
	ImapServerParams.Addr, _ = GetString("mail.server.imap.addr", ":1143")
	ImapServerParams.Domain, _ = GetString("mail.server.imap.domain", "localhost")
}
