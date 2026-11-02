// Copyright 2013 Beego Authors
// Copyright 2014 The Macaron Authors
// Copyright 2024 The Forgejo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package session

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

// FileStore represents a file session store implementation.
type FileStore struct {
	p    *FileProvider
	sid  string
	lock sync.RWMutex
	data map[any]any
}

// NewFileStore creates and returns a file session store.
func NewFileStore(p *FileProvider, sid string, kv map[any]any) *FileStore {
	return &FileStore{
		p:    p,
		sid:  sid,
		data: kv,
	}
}

// Set sets value to given key in session.
func (s *FileStore) Set(key, val any) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = val
	return nil
}

// Get gets value by given key in session.
func (s *FileStore) Get(key any) any {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data[key]
}

// Delete delete a key from session.
func (s *FileStore) Delete(key any) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
	return nil
}

// ID returns current session ID.
func (s *FileStore) ID() string {
	return s.sid
}

// Release releases resource and save data to provider.
func (s *FileStore) Release() error {
	s.p.lock.Lock()
	defer s.p.lock.Unlock()

	// Skip encoding if the data is empty
	if len(s.data) == 0 {
		return nil
	}

	data, err := EncodeGob(s.data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.p.filepath(s.sid), data, 0o600)
}

// Flush deletes all session data.
func (s *FileStore) Flush() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = make(map[any]any)
	return nil
}

// True if no keys have been set
func (s *FileStore) Empty() bool {
	return len(s.data) == 0
}

// FileProvider represents a file session provider implementation.
type FileProvider struct {
	lock        sync.RWMutex
	maxlifetime int64
	rootPath    string
}

// Init initializes file session provider with given root path.
func (p *FileProvider) Init(maxlifetime int64, rootPath string) error {
	p.lock.Lock()
	p.maxlifetime = maxlifetime
	p.rootPath = rootPath
	p.lock.Unlock()
	return nil
}

func (p *FileProvider) filepath(sid string) string {
	return path.Join(p.rootPath, string(sid[0]), string(sid[1]), sid)
}

// Read returns raw session store by session ID.
func (p *FileProvider) Read(sid string) (_ RawStore, err error) {
	filename := p.filepath(sid)
	if err = os.MkdirAll(path.Dir(filename), 0o700); err != nil {
		return nil, err
	}
	p.lock.RLock()
	defer p.lock.RUnlock()

	var f *os.File
	fStat, err := os.Stat(filename)
	if err != nil || (fStat.ModTime().Unix()+p.maxlifetime) < time.Now().Unix() {
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		f, err = os.Create(filename)
	} else {
		f, err = os.OpenFile(filename, os.O_RDONLY, 0o600)
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err = os.Chtimes(filename, time.Now(), time.Now()); err != nil {
		return nil, err
	}

	var kv map[any]any
	err = gob.NewDecoder(f).Decode(&kv)
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
		// the session file has been truncated and is now invalid - therefore all session data is lost
		kv = make(map[any]any)

	}
	return NewFileStore(p, sid, kv), nil
}

// Exist returns true if session with given ID exists.
func (p *FileProvider) Exist(sid string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	_, err := os.Stat(p.filepath(sid))
	return err == nil || os.IsExist(err)
}

// Destroy deletes a session by session ID.
func (p *FileProvider) Destroy(sid string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	return os.Remove(p.filepath(sid))
}

func (p *FileProvider) regenerate(oldsid, sid string) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	filename := p.filepath(sid)
	_, err = os.Stat(p.filepath(filename))
	if err == nil || os.IsExist(err) {
		return fmt.Errorf("new sid '%s' already exists", sid)
	}

	oldname := p.filepath(oldsid)
	fStat, err := os.Stat(oldname)
	if err != nil || fStat.IsDir() {
		data, err := EncodeGob(make(map[any]any))
		if err != nil {
			return err
		}
		if err = os.MkdirAll(path.Dir(oldname), 0o700); err != nil {
			return err
		}
		if err = os.WriteFile(oldname, data, 0o600); err != nil {
			return err
		}
	}

	if err = os.MkdirAll(path.Dir(filename), 0o700); err != nil {
		return err
	}
	if err = os.Rename(oldname, filename); err != nil {
		return err
	}
	return nil
}

// Regenerate regenerates a session store from old session ID to new one.
func (p *FileProvider) Regenerate(oldsid, sid string) (_ RawStore, err error) {
	if err := p.regenerate(oldsid, sid); err != nil {
		return nil, err
	}

	return p.Read(sid)
}

// Count counts and returns number of sessions.
func (p *FileProvider) Count() int {
	count := 0
	if err := filepath.WalkDir(p.rootPath, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			count++
		}
		return nil
	}); err != nil {
		log.Printf("error counting session files: %v", err)
		return 0
	}
	return count
}

// GC calls GC to clean expired sessions.
func (p *FileProvider) GC() {
	p.lock.RLock()
	defer p.lock.RUnlock()

	_, err := os.Stat(p.rootPath)
	if err != nil && os.IsNotExist(err) {
		return
	}

	if err := filepath.WalkDir(p.rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !d.IsDir() &&
			(info.ModTime().Unix()+p.maxlifetime) < time.Now().Unix() {
			return os.Remove(path)
		}
		return nil
	}); err != nil {
		log.Printf("error garbage collecting session files: %v", err)
	}
}

func init() {
	Register("file", &FileProvider{})
}
