// Copyright 2020 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package har

import (
	"container/list"
	"sync"
)

// EntryList implements the har.EntryContainer interface for the storage of har.Entry
type EntryList struct {
	lock  sync.Mutex
	items *list.List
}

func NewEntryList() *EntryList {
	return &EntryList{
		items: list.New(),
	}
}

// AddEntry adds an entry to the entry list
func (el *EntryList) AddEntry(entry *Entry) {
	el.lock.Lock()
	defer el.lock.Unlock()

	el.items.PushBack(entry)
}

// Entries returns a slice containing all entries
func (el *EntryList) Entries() []*Entry {
	el.lock.Lock()
	defer el.lock.Unlock()

	es := make([]*Entry, 0, el.items.Len())

	for e := el.items.Front(); e != nil; e = e.Next() {
		es = append(es, e.Value.(*Entry))
	}

	return es
}

// RemoveMatches takes a matcher function and returns all entries that return true from the function
func (el *EntryList) RemoveCompleted() []*Entry {
	el.lock.Lock()
	defer el.lock.Unlock()

	es := make([]*Entry, 0, el.items.Len())
	var next *list.Element

	for e := el.items.Front(); e != nil; e = next {
		next = e.Next()

		entry := getEntry(e)
		if entry.Response != nil {
			es = append(es, entry)
			el.items.Remove(e)
		}
	}

	return es
}

// RemoveEntry removes and entry from the entry list via the entry's id
func (el *EntryList) RemoveEntry(id string) *Entry {
	el.lock.Lock()
	defer el.lock.Unlock()

	if e, en := el.retrieveElementEntry(id); e != nil {
		el.items.Remove(e)

		return en
	}

	return nil
}

// Reset reinitializes the entrylist
func (el *EntryList) Reset() {
	el.lock.Lock()
	defer el.lock.Unlock()

	el.items.Init()
}

// RetrieveEntry returns an entry from the entrylist via the entry's id
func (el *EntryList) RetrieveEntry(id string) *Entry {
	el.lock.Lock()
	defer el.lock.Unlock()

	_, en := el.retrieveElementEntry(id)

	return en
}

func getEntry(e *list.Element) *Entry {
	if e != nil {
		return e.Value.(*Entry)
	}

	return nil
}

func (el *EntryList) retrieveElementEntry(id string) (*list.Element, *Entry) {
	for e := el.items.Front(); e != nil; e = e.Next() {
		if en := getEntry(e); en.ID == id {
			return e, en
		}
	}

	return nil, nil
}
