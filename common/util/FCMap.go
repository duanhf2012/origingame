package util

/*
这是一个固定容量的map实现，并且实现协程安全
*/
import (
	"github.com/duanhf2012/origin/v2/util/sync"
	syssync "sync"
	"time"
)

var elementPool = sync.NewPoolEx(make(chan sync.IPoolData, 10000), func() sync.IPoolData {
	return &element{}
})

// element is an element of a linked list.
type element struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *element

	// The list to which this element belongs.
	list *list

	// The value stored with this element.
	Value any
	//AppendValue1 any
	//AppendValue2 any
	ref bool
}

var empty element

func (e *element) Reset() {
	*e = empty
}

func (e *element) IsRef() bool {
	return e.ref
}

func (e *element) Ref() {
	e.ref = true
}

func (e *element) UnRef() {
	e.ref = false
}

// Next returns the next list element or nil.
func (e *element) Next() *element {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *element) Prev() *element {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// list represents a doubly linked list.
// The zero value for list is an empty list ready to use.
type list struct {
	root element // sentinel list element, only &root, root.prev, and root.next are used
	len  int     // current list length excluding (this) sentinel element
}

// Init initializes or clears list l.
func (l *list) Init() *list {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New returns an initialized list.
func New() *list { return new(list).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *list) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
func (l *list) Front() *element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *list) Back() *element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero list value.
func (l *list) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *list) insert(e, at *element) *element {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&element{Value: v}, at).
func (l *list) insertValue(v any, at *element) *element {

	ele := elementPool.Get().(*element)
	ele.Value = v

	return l.insert(ele, at)
}

// remove removes e from its list, decrements l.len
func (l *list) remove(e *element) {
	e.prev.next = e.next
	e.next.prev = e.prev

	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil

	//回收内存池
	if e != &l.root {
		elementPool.Put(e)
	}

	l.len--
}

// move moves e to next to at.
func (l *list) move(e, at *element) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *list) Remove(e *element) any {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero element) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *list) PushFront(v any) *element {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *list) PushBack(v any) *element {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *list) InsertBefore(v any, mark *element) *element {
	if mark.list != l {
		return nil
	}
	// see comment in list.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *list) InsertAfter(v any, mark *element) *element {
	if mark.list != l {
		return nil
	}
	// see comment in list.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *list) MoveToFront(e *element) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in list.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *list) MoveToBack(e *element) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in list.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *list) MoveBefore(e, mark *element) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *list) MoveAfter(e, mark *element) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *list) PushBackList(other *list) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *list) PushFrontList(other *list) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}

type FCMap struct {
	mapData     map[interface{}]*element
	pendingData *list

	syssync.RWMutex

	maxCap         int
	expirationTime int64
	lastCheckTime  int64

	checkIntervalSecond int64 //second
	intervalCheckNum    int   //检查超时个数
}

func (fm *FCMap) Init(maxCap int, expirationTime int64, checkIntervalSecond int64, intervalCheckNum int) {
	fm.maxCap = maxCap
	fm.expirationTime = expirationTime
	fm.checkIntervalSecond = checkIntervalSecond
	fm.intervalCheckNum = intervalCheckNum
	fm.mapData = make(map[interface{}]*element, maxCap)
	fm.pendingData = New()
}

type cacheData struct {
	cacheId  any
	lastTime int64
	data     any
	version  int32
}

func (fm *FCMap) UpsertData(key interface{}, Data any, version int32) {
	if fm.maxCap == 0 || fm.expirationTime == 0 {
		return
	}

	fm.Lock()
	defer fm.Unlock()

	ele, ok := fm.mapData[key]
	nowTime := time.Now().Unix()
	if ok == true {
		fm.pendingData.MoveToBack(ele)
		cData := ele.Value.(cacheData)
		cData.lastTime = nowTime
		cData.data = Data
		cData.version = version
		ele.Value = cData
	} else {
		ele = fm.pendingData.PushBack(cacheData{cacheId: key, lastTime: nowTime, data: Data, version: version})
		fm.mapData[key] = ele
	}

	//超是否超过最大量
	if fm.pendingData.Len() > fm.maxCap {
		frontEle := fm.pendingData.Front()
		if frontEle != nil {
			//log.SDebug(">++++maxCap Remove ", frontEle.Value.(cacheData).cacheId)
			delete(fm.mapData, frontEle.Value.(cacheData).cacheId)
			fm.pendingData.Remove(frontEle)
		}
	}

	//判断超时超量
	if nowTime-fm.lastCheckTime > fm.checkIntervalSecond {
		fm.lastCheckTime = nowTime
		for i := 0; i < fm.intervalCheckNum; i++ {
			frontEle := fm.pendingData.Front()
			if frontEle == nil {
				break
			}

			cdata := frontEle.Value.(cacheData)
			if nowTime-cdata.lastTime < fm.expirationTime {
				break
			}

			//log.SDebug(">++++expirationTime Remove ", cdata.cacheId)
			delete(fm.mapData, cdata.cacheId)
			fm.pendingData.Remove(frontEle)
		}
	}
}

func (fm *FCMap) FindData(key interface{}) any {
	if fm.maxCap == 0 || fm.expirationTime == 0 {
		return nil
	}

	fm.Lock()
	defer fm.Unlock()

	ele, ok := fm.mapData[key]
	if ok == false || ele == nil {
		return nil
	}

	//查找时，更新时间，并且移动到链表尾部
	data := ele.Value.(cacheData)
	data.lastTime = time.Now().Unix()
	ele.Value = data

	fm.pendingData.MoveToBack(ele)

	return ele.Value.(cacheData).data
}

func (fm *FCMap) RemoveCache(key interface{}) bool {
	fm.Lock()
	defer fm.Unlock()

	ele, ok := fm.mapData[key]
	if ok == false || ele == nil {
		return false
	}

	delete(fm.mapData, key)
	fm.pendingData.Remove(ele)

	return true
}
