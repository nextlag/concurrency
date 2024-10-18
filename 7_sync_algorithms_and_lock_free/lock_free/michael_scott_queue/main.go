package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type item struct {
	value string
	next  unsafe.Pointer
}

type Queue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func New() *Queue {
	dummy := &item{}
	return &Queue{
		head: unsafe.Pointer(dummy),
		tail: unsafe.Pointer(dummy),
	}
}

func (q *Queue) Push(value string) {
	// Создание нового узла
	node := &item{value: value}

	for {
		// Чтение указателей tail и next
		tail := atomic.LoadPointer(&q.tail)
		next := atomic.LoadPointer(&(*item)(tail).next)

		// Если tail не изменился
		if tail == atomic.LoadPointer(&q.tail) {
			// гарантия, что текущий узел последний в очереди
			if next == nil {
				// проверяем, что next все еще nil и если это так обновляем его на node
				if atomic.CompareAndSwapPointer(&(*item)(tail).next, next, unsafe.Pointer(node)) {
					// CAS был успешен, фиксируем tail
					atomic.CompareAndSwapPointer(&q.tail, tail, unsafe.Pointer(node))
					return
				}
			} else {
				// Пробуем обновить tail, если другая горутина изменила его
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			}
		}
	}
}

func (q *Queue) Pop() string {
	for {
		// Загружаем указатели head, tail и next
		head := atomic.LoadPointer(&q.head)
		tail := atomic.LoadPointer(&q.tail)
		next := atomic.LoadPointer(&(*item)(head).next)

		// очередь не пустая
		if head != tail {
			// попробуем вытащить value
			value := (*item)(next).value

			if atomic.CompareAndSwapPointer(&q.head, head, next) {
				return value
			}
		} else {
			// если голова и хвост — это какой-то узел
			if next == nil {
				// очередь содержит только фиктивный узел
				return ""
			}
			// иначе хвост надо зафиксировать
			atomic.CompareAndSwapPointer(&q.tail, tail, next)
		}
	}
}

func main() {
	queue := New()

	queue.Push("A")
	queue.Push("B")
	queue.Push("C")

	fmt.Println(queue.Pop())
	fmt.Println(queue.Pop())
	fmt.Println(queue.Pop())
	fmt.Println(queue.Pop())
	fmt.Println(queue.Pop())
}
