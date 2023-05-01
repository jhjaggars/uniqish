package compare

import "container/list"

type ListWindow struct {
	max    int
	window list.List
}

func (w *ListWindow) Add(item interface{}) {
	if w.window.Len() == w.max {
		w.window.Remove(w.window.Back())
	}
	w.window.PushFront(item)
}

func (w *ListWindow) Promote(e *list.Element) {
	w.window.MoveToFront(e)
}

func (w *ListWindow) Front() *list.Element {
	return w.window.Front()
}
