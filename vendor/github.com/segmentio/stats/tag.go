package stats

import (
	"sort"
	"sync"
)

// A Tag is a pair of a string key and value set on measures to define the
// dimensions of the metrics.
type Tag struct {
	Name  string
	Value string
}

// T is shorthand for `stats.Tag{Name: "blah", Value: "foo"}`  It returns
// the tag for Name k and Value v
func T(k, v string) Tag {
	return Tag{Name: k, Value: v}
}

func (t Tag) String() string {
	return t.Name + "=" + t.Value
}

// M allows for creating a tag list from a map.
func M(m map[string]string) []Tag {
	tags := make([]Tag, 0, len(m))
	for k, v := range m {
		tags = append(tags, T(k, v))
	}
	return tags
}

// TagsAreSorted returns true if the given list of tags is sorted by tag name,
// false otherwise.
func TagsAreSorted(tags []Tag) bool {
	if len(tags) > 1 {
		min := tags[0].Name
		for _, tag := range tags[1:] {
			if tag.Name < min {
				return false
			}
			min = tag.Name
		}
	}
	return true
}

// SortTags sorts the slice of tags.
func SortTags(tags []Tag) []Tag {
	// Insertion sort since these arrays are very small and allocation is the
	// primary enemy of performance here.
	if len(tags) >= 20 {
		sort.Sort(tagsByName(tags))
	} else {
		for i := 0; i < len(tags); i++ {
			for j := i; j > 0 && tags[j-1].Name > tags[j].Name; j-- {
				tags[j], tags[j-1] = tags[j-1], tags[j]
			}
		}
	}
	return tags
}

type tagsByName []Tag

func (t tagsByName) Len() int               { return len(t) }
func (t tagsByName) Less(i int, j int) bool { return t[i].Name < t[j].Name }
func (t tagsByName) Swap(i int, j int)      { t[i], t[j] = t[j], t[i] }

func concatTags(t1 []Tag, t2 []Tag) []Tag {
	n := len(t1) + len(t2)
	if n == 0 {
		return nil
	}
	t3 := make([]Tag, 0, n)
	t3 = append(t3, t1...)
	t3 = append(t3, t2...)
	return t3
}

func copyTags(tags []Tag) []Tag {
	if len(tags) == 0 {
		return nil
	}
	ctags := make([]Tag, len(tags))
	copy(ctags, tags)
	return ctags
}

type tagsBuffer struct {
	tags tagsByName
}

func (b *tagsBuffer) reset() {
	for i := range b.tags {
		b.tags[i] = Tag{}
	}
	b.tags = b.tags[:0]
}

func (b *tagsBuffer) sort() {
	sort.Sort(&b.tags)
}

func (b *tagsBuffer) append(tags ...Tag) {
	b.tags = append(b.tags, tags...)
}

var tagsPool = sync.Pool{
	New: func() interface{} { return &tagsBuffer{tags: make([]Tag, 0, 8)} },
}
