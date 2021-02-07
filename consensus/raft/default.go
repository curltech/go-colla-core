package raft

import (
	"github.com/curltech/go-colla-core/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/xwb1989/sqlparser"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Node struct {
	Id        raft.ServerID
	Raft      *raft.Raft
	Transport raft.Transport
}

var node *Node

/**
创建新的节点
*/
func CreateNode(id string, address string) (*Node, error) {
	//1.配置参数
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)
	opts := &hclog.LoggerOptions{}
	config.Logger = hclog.New(opts)
	config.SnapshotInterval = 20 * time.Second
	config.SnapshotThreshold = 2
	//2.FSM，有限状态机
	storeFSM := &StoreFSM{}
	//3.日志存储
	logStore, err := raftboltdb.NewBoltStore("raft-log.bolt")
	//4.节点信息存储
	stableStore, err := raftboltdb.NewBoltStore("raft-stable.bolt")
	//5.快照存储，数据
	snapshotStore, err := raft.NewFileSnapshotStore("", 1, os.Stderr)
	//6.节点通讯
	transfort, err := newRaftTransport(address)

	r, err := raft.NewRaft(config, storeFSM, logStore, stableStore, snapshotStore, transfort)

	node = &Node{Id: config.LocalID, Raft: r, Transport: transfort}

	return node, err
}

func newRaftTransport(address string) (*raft.NetworkTransport, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(tcpAddr.String(), tcpAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

/**
FSM需要实现三个方法
*/
type StoreFSM struct {
}

/**
执行一个操作日志，包含数据或者sql
*/
func (f *StoreFSM) Apply(logEntry *raft.Log) interface{} {
	reader := strings.NewReader(string(logEntry.Data))
	sqls := sqlparser.NewTokenizer(reader)
	for {
		stmt, err := sqlparser.ParseNext(sqls)
		if err == io.EOF {
			break
		}
		// execute stmt or err.
		logger.Sugar.Infof("execute sql:%v", stmt)
	}

	return nil
}

/**
产生快照
*/
func (f *StoreFSM) Snapshot() (raft.FSMSnapshot, error) {
	snapshot := &StoreSnapshot{}

	return snapshot, nil
}

/**
从快照恢复
*/
func (f *StoreFSM) Restore(serialized io.ReadCloser) error {
	return nil
}

type StoreSnapshot struct {
}

/**
产生快照的时候调用，把当前所有的数据进行序列化备份
*/
func (s *StoreSnapshot) Persist(sink raft.SnapshotSink) error {
	logger.Sugar.Infof("Snapshot Persist start!")

	return nil
}

/**
Persist操作完成后调用
*/
func (f *StoreSnapshot) Release() {
	logger.Sugar.Infof("Snapshot Release!")
}

/**
执行命令
*/
func (node *Node) Execute(cmd []byte) {
	applyFuture := node.Raft.Apply(cmd, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		logger.Sugar.Infof("raft.Apply failed:%v", err.Error())
		return
	}
}

/**
最开始启动的节点，也是leader节点
*/
func (node *Node) Start(bootstrap bool) {
	if bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      node.Id,
					Address: node.Transport.LocalAddr(),
				},
			},
		}
		node.Raft.BootstrapCluster(configuration)
	} else {
		node.ApplyJoin("")
	}
}

/**
发送申请加入的请求
*/
func (node *Node) ApplyJoin(address string) {

}

/**
先启动的节点收到申请加入的请求后，加一个新的节点，参数为新节点的地址
*/
func (node *Node) Join(address string) {
	addPeerFuture := node.Raft.AddVoter(raft.ServerID(address),
		raft.ServerAddress(address),
		0, 0)
	if err := addPeerFuture.Error(); err != nil {
		logger.Sugar.Infof("Error joining peer to raft, address:%s, err:%v, code:%d", address, err, http.StatusInternalServerError)

		return
	}
}
