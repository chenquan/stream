/*
 *
 *     Copyright 2020 chenquan
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
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func equal(t *testing.T, stream *Stream, data []interface{}) {
	items := make([]interface{}, 0)
	for item := range stream.source {
		items = append(items, item)
	}
	if !reflect.DeepEqual(items, data) {
		t.Errorf(" %v, want %v", items, data)
	}
}
func assetEqual(t *testing.T, except interface{}, data interface{}) {
	if !reflect.DeepEqual(except, data) {
		t.Errorf(" %v, want %v", data, except)
	}

}
func TestEmpty(t *testing.T) {
	empty := Empty()
	assetEqual(t, len(empty.source), 0)
	assetEqual(t, cap(empty.source), 0)
	empty.Foreach(func(item interface{}) {

	})
}
func TestRange(t *testing.T) {
	stream1 := Range(make(chan interface{}))
	assetEqual(t, len(stream1.source), 0)

	stream2 := Range(make(chan interface{}, 2))
	assetEqual(t, len(stream2.source), 0)
	assetEqual(t, cap(stream2.source), 2)
}
func TestOf(t *testing.T) {
	ints := []interface{}{1, 2, 3, 4}
	of := Of(ints...).Sort(func(a, b interface{}) bool {
		return a.(int) < b.(int)
	})
	assetEqual(t, len(of.source), 4)
	assetEqual(t, cap(of.source), 4)
	var items []interface{}
	for item := range of.source {
		items = append(items, item)
	}
	assetEqual(t, items, ints)
}

func TestConcat(t *testing.T) {
	a1 := []interface{}{1, 2, 3}
	a2 := []interface{}{4, 5, 6}
	s1 := Of(a1...)
	s2 := Of(a2...)
	stream := Concat(s1, s2)
	var items []interface{}
	for item := range stream.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(int) < items[j].(int)
	})
	ints := make([]interface{}, 0)
	ints = append(ints, a1...)
	ints = append(ints, a2...)
	assetEqual(t, ints, items)

}
func TestFrom(t *testing.T) {
	ints := make([]interface{}, 0)
	stream := From(func(source chan<- interface{}) {
		for i := 0; i < 10; i++ {
			source <- i
			ints = append(ints, i)
		}
	})
	items := make([]interface{}, 0)
	for item := range stream.source {
		items = append(items, item)
	}
	assetEqual(t, items, ints)
}
func TestStream_Distinct(t *testing.T) {
	stream := Of(1, 2, 3, 4, 4, 22, 2, 1, 4).Distinct(func(item interface{}) interface{} {
		return item
	})
	equal(t, stream, []interface{}{1, 2, 3, 4, 22})
}
func TestStream_Count(t *testing.T) {
	data := []interface{}{1, 2, 3, 4, 4, 22, 2, 1, 4}
	assetEqual(t, Of(data...).Count(), len(data))
}

func TestStream_Buffer(t *testing.T) {
	stream := Of(1, 2, 4)
	assetEqual(t, len(stream.source), 3)
	assetEqual(t, cap(stream.source), 3)
	stream = stream.Buffer(10)
	assetEqual(t, cap(stream.source), 10)
}
func TestStream_Finish(t *testing.T) {
	var items1 []interface{}
	var items2 []interface{}
	Of(1, 2, 4).Finish(
		func(item interface{}) {
			items1 = append(items1, item)
		},
		func(item interface{}) {
			items2 = append(items2, item)
		},
	)
	assetEqual(t, items1, []interface{}{1, 2, 4})
	assetEqual(t, items2, []interface{}{1, 2, 4})
}
func TestStream_Chan(t *testing.T) {
	var items []interface{}

	for item := range Of(1, 2, 3).Chan() {
		items = append(items, item)
	}
	assetEqual(t, items, []interface{}{1, 2, 3})
}
func TestStream_Split(t *testing.T) {

	stream := Of(1, 2, 444, 441, 1).Split(3)
	assetEqual(t, (<-stream.source).([]interface{}), []interface{}{1, 2, 444})
	assetEqual(t, (<-stream.source).([]interface{}), []interface{}{441, 1})
}
func TestStream_SplitSteam2(t *testing.T) {
	streams := Of(1, 2, 444, 441, 1).SplitSteam(3)

	equal(t, (<-streams.source).(*Stream), []interface{}{1, 2, 444})
	equal(t, (<-streams.source).(*Stream), []interface{}{441, 1})
}

func TestStream_Sort(t *testing.T) {
	ints := []interface{}{4, 2, 1, 441, 23, 14, 1, 23}
	stream := Of(ints...).Sort(func(a, b interface{}) bool {
		return a.(int) < b.(int)
	})
	sort.Slice(ints, func(i, j int) bool {
		return ints[i].(int) < ints[j].(int)
	})
	equal(t, stream, ints)
}
func TestStream_Tail(t *testing.T) {
	equal(t, Of(1, 232, 3, 2, 3).Tail(1), []interface{}{3})
	equal(t, Of(1, 232, 3, 2, 3).Tail(2), []interface{}{2, 3})
}

func TestStream_Head(t *testing.T) {
	equal(t, Of(1, 232, 3, 2, 3).Head(2), []interface{}{1, 232})
	equal(t, Of(1, 232, 3, 2, 3).Head(1), []interface{}{1})

}
func TestStream_Skip(t *testing.T) {
	assetEqual(t, 3, Of(1, 2, 3, 4).Skip(1).Count())
	assetEqual(t, 1, Of(1, 2, 3, 4).Skip(3).Count())
	equal(t, Of(1, 2, 3, 4).Skip(3), []interface{}{4})
}
func TestStream_Limit(t *testing.T) {

	equal(t, Of(1, 2, 3, 4).Limit(3), []interface{}{1, 2, 3})
	equal(t, Of(1, 2, 3, 4).Limit(4), []interface{}{1, 2, 3, 4})
	equal(t, Of(1, 2, 3, 4).Limit(5), []interface{}{1, 2, 3, 4})
}
func TestStream_Foreach(t *testing.T) {
	var items []interface{}
	Of(1, 2, 3, 4).Foreach(func(item interface{}) {
		items = append(items, item)
	})
	assetEqual(t, []interface{}{1, 2, 3, 4}, items)
}
func TestStream_ForeachOrdered(t *testing.T) {
	var items []interface{}
	Of(1, 2, 3, 4).ForeachOrdered(func(item interface{}) {
		items = append(items, item)
	})
	assetEqual(t, []interface{}{4, 3, 2, 1}, items)
}
func TestStream_Concat(t *testing.T) {
	stream := Of(1).Concat(Of(2), Of(3))
	var items []interface{}
	for item := range stream.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(int) < items[j].(int)
	})
	assetEqual(t, []interface{}{1, 2, 3}, items)
}

func TestStream_Filter(t *testing.T) {
	equal(t, Of(1, 2, 3, 4).Filter(func(item interface{}) bool {
		return item.(int) > 3
	}), []interface{}{4})
	equal(t, Of(1, 2, 3, 4).Filter(func(item interface{}) bool {
		return item.(int) > 2
	}).Sort(func(a, b interface{}) bool {
		return a.(int) < b.(int)
	}), []interface{}{3, 4})
}
func TestStream_Map(t *testing.T) {
	equal(t, Of(1, 2, 3).Map(func(item interface{}) interface{} {
		return item.(int) + 1
	}).Sort(func(a, b interface{}) bool {
		return a.(int) < b.(int)
	}), []interface{}{2, 3, 4})
}
func TestStream_FlatMap2(t *testing.T) {
	equal(t,
		Of([]interface{}{1, 2}, []interface{}{3, 4}).FlatMap(func(item interface{}) interface{} {
			return item
		}).Sort(func(a, b interface{}) bool {
			return a.(int) < b.(int)
		}),
		[]interface{}{1, 2, 3, 4},
	)
}
func TestStream_Group(t *testing.T) {
	equal(t,
		Of(1, 2, 3, 4).Group(func(item interface{}) interface{} {
			return item.(int) % 2
		}).Map(func(item interface{}) interface{} {
			m := item.([]interface{})
			var vs []interface{}
			for _, v := range m {
				vs = append(vs, v)
			}
			return m
		}).FlatMap(func(item interface{}) interface{} {
			return item
		}).Sort(func(a, b interface{}) bool {
			return a.(int) < b.(int)
		}),
		[]interface{}{1, 2, 3, 4},
	)
}
func TestStream_Merge(t *testing.T) {
	equal(t, Of(1, 2, 3, 4).Merge(), []interface{}{

		[]interface{}{1, 2, 3, 4},
	})
}
func TestStream_Reverse(t *testing.T) {
	equal(t, Of(1, 2, 3, 4, 1).Reverse(), []interface{}{1, 4, 3, 2, 1})
}

func TestStream_FlatMap(t *testing.T) {
	Of([]interface{}{1, 2}, []interface{}{4, 5}).FlatMap(func(item interface{}) interface{} {
		return item
	}).Foreach(func(item interface{}) {
		fmt.Println(item)
	})
}
func TestStream_ParallelFinish(t *testing.T) {

	Of(1, 23).ParallelFinish(func(item interface{}) {

	}, WithWorkSize(2))
}

func TestStream_AnyMach(t *testing.T) {
	assetEqual(t, false, Of(1, 2, 3).AnyMach(func(item interface{}) bool {
		return 4 == item.(int)
	}))
	assetEqual(t, true, Of(1, 2, 3).AnyMach(func(item interface{}) bool {
		return 2 == item.(int)
	}))
}
func TestStream_AllMach(t *testing.T) {
	assetEqual(
		t, true, Of(1, 2, 3).AllMach(func(item interface{}) bool {
			return true
		}),
	)
	assetEqual(
		t, false, Of(1, 2, 3).AllMach(func(item interface{}) bool {
			return false
		}),
	)
	assetEqual(
		t, false, Of(1, 2, 3).AllMach(func(item interface{}) bool {
			return item.(int) == 1
		}),
	)
}

func TestStream_String(t *testing.T) {
	assetEqual(t, fmt.Sprintf("Stream{len:%d,cap:%d}", 1, 1), Of(1).String())
}
