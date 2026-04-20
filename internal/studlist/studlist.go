package studlist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const apiURL = "https://www5.pyc.edu.hk/pycnet/api/stud_list.php"

type Student struct {
	PYCCode   string `json:"pyccode"`
	Cls       string `json:"cls"`
	Num       string `json:"num"`
	UserEName string `json:"userename"`
	UserCName string `json:"usercname"`
}

var (
	mu        sync.RWMutex
	cache     map[string]*Student
	fetchedAt time.Time
	ttl       = 10 * time.Minute
)

// Lookup returns the student for the given pyccode, or nil if not found.
func Lookup(pyccode string) *Student {
	mu.RLock()
	if cache != nil && time.Since(fetchedAt) < ttl {
		s := cache[pyccode]
		mu.RUnlock()
		return s
	}
	mu.RUnlock()
	refresh()
	mu.RLock()
	defer mu.RUnlock()
	return cache[pyccode]
}

// DisplayName returns "CLS NUM ENAME" for students, or pyccode for teachers/unknown.
func DisplayName(pyccode string) string {
	s := Lookup(pyccode)
	if s == nil {
		return pyccode
	}
	return fmt.Sprintf("%s %s %s", s.Cls, s.Num, s.UserEName)
}

func refresh() {
	mu.Lock()
	defer mu.Unlock()
	// Double-check after acquiring write lock
	if cache != nil && time.Since(fetchedAt) < ttl {
		return
	}
	resp, err := http.Get(apiURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var students []Student
	if err := json.NewDecoder(resp.Body).Decode(&students); err != nil {
		return
	}
	m := make(map[string]*Student, len(students))
	for i := range students {
		m[students[i].PYCCode] = &students[i]
	}
	cache = m
	fetchedAt = time.Now()
}
