/*
 *
 *     Copyright 2021 chenquan
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

package stream

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type (
	KeyFunc  func(item interface{}) interface{}
	LessFunc func(a, b interface{}) bool

	Stream struct {
		source <-chan interface{}
	}
)

var empty = &Stream{source: make(chan interface{})}

func (s *Stream) String() string {
	return fmt.Sprintf("Stream{len:%d,cap:%d}", len(s.source), cap(s.source))
}

type MapFunc func(item interface{}) interface{}

func Empty() *Stream {
	return empty
}
func Range(source <-chan interface{}) *Stream {
	return &Stream{
		source: source,
	}
}
func Of(items ...interface{}) *Stream {
	source := make(chan interface{}, len(items))
	for _, item := range items {
		source <- item
	}
	close(source)
	return Range(source)
}
func Concat(a *Stream, others ...*Stream) *Stream {
	return a.Concat(others...)
}

func From(generate func(source chan<- interface{})) *Stream {
	source := make(chan interface{})

	NewGoroutine(func() {
		defer close(source)
		generate(source)
	})
	return Range(source)
}

func (s *Stream) Distinct(f KeyFunc) *Stream {
	source := make(chan interface{})

	NewGoroutine(func() {
		defer close(source)
		unique := make(map[interface{}]struct{})
		for item := range s.source {
			k := f(item)
			if _, ok := unique[k]; !ok {
				source <- item
				unique[k] = struct{}{}
			}
		}
	})
	return Range(source)
}
func (s *Stream) Count() (count int) {
	for range s.source {
		count++
	}
	return
}
func (s *Stream) Buffer(n int) *Stream {
	if n < 0 {
		n = 0
	}
	source := make(chan interface{}, n)
	go func() {
		for item := range s.source {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}
func (s *Stream) Finish(fs ...func(item interface{})) {
	for item := range s.source {
		for _, f := range fs {
			f(item)
		}
	}
}
func (s *Stream) Chan() <-chan interface{} {
	return s.source
}
func (s *Stream) Split(n int) *Stream {
	if n < 1 {
		panic("n should be greater than 0")
	}
	source := make(chan interface{})
	go func() {
		var chunk []interface{}
		for item := range s.source {
			chunk = append(chunk, item)
			if len(chunk) == n {
				source <- chunk
				chunk = nil
			}
		}
		if chunk != nil {
			source <- chunk
		}
		close(source)
	}()
	return Range(source)
}
func (s *Stream) SplitSteam(n int) *Stream {
	if n < 1 {
		panic("n should be greater than 0")
	}
	source := make(chan interface{})

	var chunkSource = make(chan interface{}, n)
	go func() {

		for item := range s.source {
			chunkSource <- item
			if len(chunkSource) == n {

				source <- Range(chunkSource)
				close(chunkSource)
				chunkSource = nil
				chunkSource = make(chan interface{}, n)
			}
		}
		if len(chunkSource) != 0 {
			source <- Range(chunkSource)
			close(chunkSource)
		}
		close(source)
	}()
	return Range(source)
}

func (s *Stream) Sort(less LessFunc) *Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})
	return Of(items...)
}
func (s *Stream) Tail(n int64) *Stream {
	if n < 1 {
		panic("n should be greater than 0")
	}
	source := make(chan interface{})

	go func() {
		ring := NewRing(int(n))
		for item := range s.source {
			ring.Add(item)
		}
		for _, item := range ring.Take() {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

func (s *Stream) Head(n int64) *Stream {
	if n < 1 {
		panic("n must be greater than 0")
	}

	source := make(chan interface{})

	go func() {
		for item := range s.source {
			n--
			if n >= 0 {
				source <- item
			}
			if n == 0 {
				// let successive method go ASAP even we have more items to skip
				// why we don't just break the loop, because if break,
				// this former goroutine will block forever, which will cause goroutine leak.
				close(source)
			}
		}
		if n > 0 {
			close(source)
		}
	}()

	return Range(source)
}

func (s *Stream) Skip(size int) *Stream {
	if size == 0 {
		return s
	}
	if size < 0 {
		panic("size must be greater than -1")
	}
	source := make(chan interface{})

	go func() {
		i := 0
		for item := range s.source {
			if i >= size {
				source <- item
			}
			i++
		}
		close(source)
	}()
	return Range(source)
}

func (s *Stream) Limit(size int) *Stream {
	if size == 0 {
		return Of()
	}
	if size < 0 {
		panic("size must be greater than -1")
	}
	source := make(chan interface{})

	go func() {
		i := 0
		for item := range s.source {
			if i != size {
				source <- item
			}
			i++
		}
		close(source)
	}()
	return Range(source)
}

func (s *Stream) Foreach(f func(item interface{})) {
	items := make([]interface{}, 0)
	for item := range s.source {
		f(item)
		items = append(items, item)
	}

}
func (s *Stream) ForeachOrdered(f func(item interface{})) {
	items := make([]interface{}, 0)
	for item := range s.source {
		items = append(items, item)
	}
	n := len(items)
	for i := 0; i < n; i++ {
		f(items[n-i-1])
	}

}

func (s *Stream) Concat(others ...*Stream) *Stream {
	source := make(chan interface{})
	wg := sync.WaitGroup{}
	for _, other := range others {
		if s == other {
			panic("other same")
		}
		wg.Add(1)
		go func(iother *Stream) {
			for item := range iother.source {
				source <- item
			}
			wg.Done()
		}(other)

	}

	wg.Add(1)
	go func() {
		for item := range s.source {
			source <- item
		}
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(source)
	}()
	return Range(source)
}

type FilterFunc func(item interface{}) bool

func (s *Stream) Filter(fn FilterFunc, opts ...Option) *Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		if fn(item) {
			pipe <- item
		}
	}, opts...)
}

func (s *Stream) Walk(f func(item interface{}, pipe chan<- interface{}), opts ...Option) *Stream {
	option := loadOptions(opts...)
	pipe := make(chan interface{}, option.workSize)
	go func() {
		var wg sync.WaitGroup
		pool := make(chan struct{}, option.workSize)

		for {
			pool <- struct{}{}
			item, ok := <-s.source
			if !ok {
				<-pool
				break
			}

			wg.Add(1)
			// better to safely run caller defined method
			NewGoroutine(func() {
				defer func() {
					wg.Done()
					<-pool
				}()
				f(item, pipe)
			})
		}
		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

type WalkFunc func(item interface{}, pipe chan<- interface{})

func (s *Stream) Map(fn MapFunc, opts ...Option) *Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		pipe <- fn(item)
	}, opts...)
}

func (s *Stream) FlatMap(fn MapFunc, opts ...Option) *Stream {
	return s.Walk(func(item interface{}, pipe chan<- interface{}) {
		switch v := item.(type) {
		case []interface{}:
			for _, x := range v {
				pipe <- fn(x)
			}
		case interface{}:
			pipe <- fn(v)
		}
	}, opts...)
}

func (s *Stream) Group(f KeyFunc) *Stream {
	groups := make(map[interface{}][]interface{})
	for item := range s.source {
		key := f(item)
		groups[key] = append(groups[key], item)
	}

	source := make(chan interface{})
	go func() {
		for _, group := range groups {
			source <- group
		}
		close(source)
	}()

	return Range(source)
}
func (s *Stream) Merge() *Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}

	source := make(chan interface{}, 1)
	source <- items
	close(source)

	return Range(source)
}
func (s *Stream) Reverse() *Stream {
	var items []interface{}
	for item := range s.source {
		items = append(items, item)
	}
	for i := len(items)/2 - 1; i >= 0; i-- {
		opp := len(items) - 1 - i
		items[i], items[opp] = items[opp], items[i]
	}
	return Of(items...)
}

func (s *Stream) ParallelFinish(fn func(item interface{}), opts ...Option) {
	s.Walk(func(item interface{}, pipe chan<- interface{}) {
		fn(item)
	}, opts...).Finish()
}
func (s *Stream) AnyMach(f func(item interface{}) bool, opts ...Option) (isFind bool) {
	for item := range s.source {
		if f(item) {
			isFind = true
		}
	}
	return
}
func (s *Stream) AllMach(f func(item interface{}) bool, opts ...Option) (isFind bool) {
	isFind = true
	for item := range s.source {
		if !f(item) {
			isFind = false
			return
		}
	}
	return
}
func (s *Stream) findFirst() (result interface{}, err error) {
	for item := range s.source {
		result = item
	}
	if result == nil {
		err = errors.New("no element")
	}
	return
}
func (s *Stream) Peek(f func(item interface{})) *Stream {
	source := make(chan interface{})
	go func() {
		for item := range s.source {
			source <- item
			f(item)
		}
		close(source)
	}()
	return Range(source)
}
