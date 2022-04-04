package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	first *ListItem
	last  *ListItem
}

func NewList() List {
	return &list{
		len:   0,
		first: nil,
		last:  nil,
	}
}

func (l list) Len() int {
	return l.len
}

func (l list) Front() *ListItem {
	return l.first
}

func (l list) Back() *ListItem {
	return l.last
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := ListItem{
		Value: v,
		Next:  l.first,
	}
	if l.first != nil {
		l.first.Prev = &newItem
	} else {
		l.last = &newItem
	}
	l.first = &newItem
	l.len++
	return &newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	newItem := ListItem{
		Value: v,
		Prev:  l.last,
	}
	if l.last != nil {
		l.last.Next = &newItem
	} else {
		l.first = &newItem
	}
	l.last = &newItem
	l.len++
	return &newItem
}

func (l *list) Remove(i *ListItem) {
	if l.len == 1 {
		l.first = nil
		l.last = nil
	} else {
		if i.Next == nil { // это последний элт списка
			l.last = i.Prev
			i.Prev.Next = nil
		} else if i.Prev == nil { // первый элт списка
			l.first = i.Next
			i.Next.Prev = nil
		} else {
			i.Next.Prev = i.Prev
			i.Prev.Next = i.Next
		}
	}
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if l.first != i {
		l.Remove(i) // технически мы удаляем элемент и вставляем спереди
		l.len++     // тут идет "вставка спереди"
		l.first.Prev = i
		i.Prev = nil
		i.Next = l.first
		l.first = i
	}
}
