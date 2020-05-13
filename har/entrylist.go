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

type EntryList struct {
	lock  sync.Mutex
	Items *list.List
}

func NewEntryList() *EntryList {
	return &EntryList{
		Items: list.New(),
	}
}

func (el *EntryList) Lock() {
	el.lock.Lock()
}

func (el *EntryList) Unlock() {
	el.lock.Unlock()
}

// AddEntry adds an entry to the entry list
func (el *EntryList) AddEntry(entry *Entry) {
	el.Lock()
	defer el.Unlock()

	el.Items.PushBack(entry)
}

// Entries returns a slice containing all entries
func (el *EntryList) Entries() []*Entry {
	el.Lock()
	defer el.Unlock()

	es := make([]*Entry, 0, el.Items.Len())

	for e := el.Items.Front(); e != nil; e = e.Next() {
		es = append(es, e.Value.(*Entry))
	}

	return es
}

// RemoveMatches takes a matcher function and returns all entries that return true from the function
func (el *EntryList) RemoveMatches(matcher func(*Entry) bool) []*Entry {
	el.Lock()
	defer el.Unlock()

	es := make([]*Entry, 0, el.Items.Len())
	var next *list.Element

	for e := el.Items.Front(); e != nil; e = next {
		next = e.Next()

		entry := getEntry(e)
		if matcher(entry) {
			es = append(es, entry)
			el.Items.Remove(e)
		}
	}

	return es
}

// RemoveEntry removes and entry from the entry list via the entry's id
func (el *EntryList) RemoveEntry(id string) *Entry {
	el.Lock()
	defer el.Unlock()

	if e, en := el.retrieveElementEntry(id); e != nil {
		el.Items.Remove(e)

		return en
	}

	return nil
}

// Reset reinitializes the entrylist
func (el *EntryList) Reset() {
	el.Lock()
	defer el.Unlock()

	el.Items.Init()
}

// RetrieveEntry returns an entry from the entrylist via the entry's id
func (el *EntryList) RetrieveEntry(id string) *Entry {
	el.Lock()
	defer el.Unlock()

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
	for e := el.Items.Front(); e != nil; e = e.Next() {
		if en := getEntry(e); en.ID == id {
			return e, en
		}
	}

	return nil, nil
}
