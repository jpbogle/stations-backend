package utils

import (
	"container/list"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Session interface {
	Set(key string, value interface{}) //set session value
	Get(key string) interface{}        //get session value
	Delete(key string)                 //delete session value
	SessionID() string                 //back current sessionID
}

type sessionManager struct {
	maxlifetime int64
	cookieName  string                   //private cookiename
	lock        sync.Mutex               // protects session
	sessions    map[string]*list.Element // save in memory
	list        *list.List               // gc
}
type session struct {
	sid          string                 // unique session id
	timeAccessed time.Time              // last access time
	value        map[string]interface{} // session value stored inside
}

// TODO: lock cookie down to domain
const cookieLifetime = 60 * 60 * 24
const cookieName = "stations.session"

var globalSessions = CreateSessionManager() // one day

func CreateSessionManager() *sessionManager {
	manager := &sessionManager{
		cookieName:  cookieName,
		maxlifetime: cookieLifetime,
		sessions:    make(map[string]*list.Element, 0),
		list:        list.New(),
	}
	go time.AfterFunc(time.Second, manager.gc)
	return manager
}

//////////////////////////////////////////////////////////////////////////////
//
// Public SessionManager methods
//
//////////////////////////////////////////////////////////////////////////////

// Returns a the existing Session for the client or creates a new Session
func (sm *sessionManager) GetSession(ctx *Context) Session {
	cookie, err := ctx.Req.Cookie(sm.cookieName)
	if err != nil || cookie.Value == "" {
		sid := sm.createRandomId()
		session := sm.createSession(sid)
		cookie := http.Cookie{
			Name:     sm.cookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/api",
			HttpOnly: true,
			MaxAge:   int(sm.maxlifetime),
		}
		http.SetCookie(ctx.Res, &cookie)
		return session
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	session := sm.readSession(sid)
	return session
}

// Destroys the client's existing session
func (sm *sessionManager) DestroySession(ctx *Context) {
	cookie, err := ctx.Req.Cookie(sm.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	sm.destroySession(cookie.Value)
	expiration := time.Now()
	cookie = &http.Cookie{
		Name:     sm.cookieName,
		Path:     "/api",
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1,
	}
	http.SetCookie(ctx.Res, cookie)
}

//////////////////////////////////////////////////////////////////////////////
//
// Private *sessionManager methods
//
//////////////////////////////////////////////////////////////////////////////

// Creates a random id for a new session
func (sm *sessionManager) createRandomId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// Creates a new session with the given session id
func (sm *sessionManager) createSession(sid string) Session {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	v := make(map[string]interface{}, 0)
	newsess := &session{
		sid:          sid,
		timeAccessed: time.Now(),
		value:        v,
	}
	element := sm.list.PushBack(newsess)
	sm.sessions[sid] = element
	return newsess
}

// Fetches the session with the given session id, or creates a new session with the given session id
func (sm *sessionManager) readSession(sid string) Session {
	if element, ok := sm.sessions[sid]; ok {
		return element.Value.(*session)
	}
	sess := sm.createSession(sid)
	return sess
}

// Deletes the session with the given session id
func (sm *sessionManager) destroySession(sid string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if element, ok := sm.sessions[sid]; ok {
		delete(sm.sessions, sid)
		sm.list.Remove(element)
	}
}

// Updates the timeAccessed property of the session with the given session id
func (sm *sessionManager) updateSession(sid string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if element, ok := sm.sessions[sid]; ok {
		element.Value.(*session).timeAccessed = time.Now()
		sm.list.MoveToFront(element)
	}
}

// Deletes all expired sessions and reschedules another call to gc();
// Only to be called in createSessionManager
func (sm *sessionManager) gc() {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for {
		element := sm.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*session).timeAccessed.Unix() + sm.maxlifetime) < time.Now().Unix() {
			sm.list.Remove(element)
			delete(sm.sessions, element.Value.(*session).sid)
		} else {
			break
		}
	}
	time.AfterFunc(time.Duration(sm.maxlifetime)*time.Millisecond, sm.gc)
}

//////////////////////////////////////////////////////////////////////////////
//
// Session methods
//
//////////////////////////////////////////////////////////////////////////////

// Set a value in the session
func (s *session) Set(key string, value interface{}) {
	s.value[key] = value
	globalSessions.updateSession(s.sid)
}

// Get a value in the session
func (s *session) Get(key string) interface{} {
	globalSessions.updateSession(s.sid)
	if v, ok := s.value[key]; ok {
		return v
	}
	return nil
}

// Delete a value in the session
func (s *session) Delete(key string) {
	delete(s.value, key)
	globalSessions.updateSession(s.sid)
}

// Get the session's id
func (s *session) SessionID() string {
	return s.sid
}
