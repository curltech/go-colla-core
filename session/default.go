package session

import (
	"container/list"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/curltech/go-colla-core/logger"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

/**
一个系统有多个会话池，每个会话池有自己的名称
每个会话池内有多个会话，每个会话有自己的ID作为主键,和一系列的key-value对
*/

type Session struct {
	sid          string //session id唯一标示
	isNew        bool
	timeAccessed time.Time                   //最后访问时间
	value        map[interface{}]interface{} //session里面存储的值
}

/*
*
Set(key, value interface{}) error //设置Session
Get(key interface{}) interface{}  //获取Session
Delete(key interface{}) error     //删除Session
SessionID() string                //当前SessionID
*/
func (this *Session) Set(key, value interface{}) error {
	this.value[key] = value
	sessionManager.sessionPool.Update(this.sid)
	return nil
}
func (this *Session) Get(key interface{}) interface{} {
	sessionManager.sessionPool.Update(this.sid)
	if v, ok := this.value[key]; ok {
		return v
	} else {
		return nil
	}
}
func (this *Session) Delete(key interface{}) error {
	delete(this.value, key)
	sessionManager.sessionPool.Update(this.sid)
	return nil
}
func (this *Session) SessionID() string {
	return this.sid
}
func (this *Session) IsNew() bool {
	return this.isNew
}

/*
*
会话池用于存放连接上来的多个会话，会话管理器Start的时候如果是新会话，将放入会话池
*/
type SessionPool struct {
	lock     sync.Mutex               //用来锁
	sessions map[string]*list.Element //用来存储在内存
	list     *list.List               //用来做gc
}

/*
*
SessionInit(sid string) (Session, error)
SessionRead(sid string) (Session, error)
SessionDestroy(sid string) error
SessionGC(maxLifeTime int64)
*/
func (this *SessionPool) Init(sid string) (*Session, error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &Session{sid: sid, isNew: true, timeAccessed: time.Now(), value: v}
	element := this.list.PushBack(newsess)
	this.sessions[sid] = element

	return newsess, nil
}
func (this *SessionPool) Read(sid string) (*Session, error) {
	if element, ok := this.sessions[sid]; ok {
		return element.Value.(*Session), nil
	} else {
		sess, err := this.Init(sid)
		return sess, err
	}
	return nil, nil
}
func (this *SessionPool) Destroy(sid string) error {
	if element, ok := this.sessions[sid]; ok {
		delete(this.sessions, sid)
		this.list.Remove(element)
		return nil
	}
	return nil
}
func (this *SessionPool) SessionGC(maxlifetime int64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for {
		element := this.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*Session).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			sessionManager.sessionPool.list.Remove(element)
			delete(sessionManager.sessionPool.sessions, element.Value.(*Session).sid)
		} else {
			break
		}
	}
}
func (this *SessionPool) Update(sid string) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if element, ok := this.sessions[sid]; ok {
		element.Value.(*Session).isNew = false
		element.Value.(*Session).timeAccessed = time.Now()
		this.list.MoveToFront(element)
		return nil
	}
	return nil
}

var sessionPools = make(map[string]*SessionPool)

/*
*
会话管理器获取当前会话，GetDefault().Start()，从而进一步获取会话id，会话变量的存取
*/
type SessionManager struct {
	cookieName  string       //cookie的名称
	lock        sync.Mutex   //锁，保证并发时数据的安全一致
	sessionPool *SessionPool //管理session
	maxLifeTime int64        //超时时间
}

var sessionManager *SessionManager

func NewSessionManager(poolName, cookieName string, maxLifetime int64) (*SessionManager, error) {
	sessionPool, p := sessionPools[poolName]
	if p {
		return nil, errors.New("session: SessionManager is existed")
	} else {
		sessionPool = &SessionPool{list: list.New(), sessions: make(map[string]*list.Element, 0)}
	}

	//返回一个 Manager 对象
	sessionManager = &SessionManager{
		cookieName:  cookieName,
		maxLifeTime: maxLifetime,
		sessionPool: sessionPool,
	}
	go sessionManager.SessionGC()

	return sessionManager, nil
}

func GetDefault() *SessionManager {
	return sessionManager
}

func init() {
	NewSessionManager("memory", "customsessionid", 3600)
}

func (manager *SessionManager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 根据当前请求的cookie中判断是否存在有效的session, 不存在则创建
func (this *SessionManager) Start(w http.ResponseWriter, r *http.Request) (session *Session) {
	//为该方法加锁
	this.lock.Lock()
	defer this.lock.Unlock()
	//获取 request 请求中的 cookie 值
	cookie, err := r.Cookie(this.cookieName)
	if err != nil || cookie.Value == "" {
		sid := this.sessionId()
		logger.Sugar.Infof("New session:%v", sid)
		session, _ = this.sessionPool.Init(sid)
		cookie := http.Cookie{
			Name:     this.cookieName,
			Value:    url.QueryEscape(sid), //转义特殊符号@#￥%+*-等
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(this.maxLifeTime)}

		http.SetCookie(w, &cookie) //将新的cookie设置到响应中
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		logger.Sugar.Infof("Get old session:%v", sid)
		session, _ = this.sessionPool.Read(sid)
	}
	return
}

// SessionDestroy 注销 Session
func (this *SessionManager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(this.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	this.sessionPool.Destroy(cookie.Value)
	expiredTime := time.Now()
	newCookie := http.Cookie{
		Name: this.cookieName,
		Path: "/", HttpOnly: true,
		Expires: expiredTime,
		MaxAge:  -1, //会话级cookie
	}
	http.SetCookie(w, &newCookie)
}

// 记录该session被访问的次数
func test(w http.ResponseWriter, r *http.Request) {
	sess := sessionManager.Start(w, r)   //获取session实例
	createTime := sess.Get("createTime") //获得该session的创建时间
	if createTime == nil {
		sess.Set("createTime", time.Now().Unix())
	} else if (createTime.(int64) + 360) < (time.Now().Unix()) { //已过期
		//注销旧的session信息，并新建一个session  globalSession.SessionDestroy(w, r)
		sess = sessionManager.Start(w, r)
	}
	count := sess.Get("countnum")
	if count == nil {
		sess.Set("countnum", 1)
	} else {
		sess.Set("countnum", count.(int)+1)
	}
}

func (this *SessionManager) SessionGC() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.sessionPool.SessionGC(this.maxLifeTime)
	//使用time包中的计时器功能，它会在session超时时自动调用GC方法
	time.AfterFunc(time.Duration(this.maxLifeTime), func() {
		this.SessionGC()
	})
}
