package timeline

// NOTE: вся эта хуйня сгенерирована нейронкой
// я не уверен что она мне вообще подходит, но для начала пойдет
// кароче да, он даже в оригинальной версии неправильно работает, буду переделывать
// UPD: все норм, теперь вроде работает
// кстати, timestamp можно ставить любой, тоесть так можно реализоввыать задержку, но это уже другая история.

// TODO: кейс с переполнением канала
// почему небуферезированнй канал?
// добавить Flush для закрытия
//
// как использовать:
//
// Создаем timeline на 1000 событий
// tl := timeline.NewTimeline(1000)
//
// Добавляем события
// tl.AddEvent(timeline.EventChat, "привет всем!", "user123")
// tl.AddEvent(timeline.EventSpeech, "стример сказал что-то важное")
// tl.AddEvent(timeline.EventScreenshot, "описание скриншота от LLAVA")
//
// Получаем данные для AI
// recent := tl.GetRecentEvents(5 * time.Minute)  // за последние 5 минут
// last50 := tl.GetLastEvents(50)                 // последние 50 событий
//

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/godovasik/dawgobot/logger"
)

// Типы событий
type EventType int

const (
	EventGlobal EventType = iota
	EventChat
	EventImage
	EventSpeech
	EventScreenshot
)

// Структура события
type Event struct {
	Type      EventType
	Content   string
	Author    string // для чата
	Streamer  string
	Timestamp time.Time
}

// Циркулярный буфер
type CircularBuffer struct {
	events []Event
	size   int
	head   int // где начинается актуальная часть
	tail   int // куда записываем следующий
	count  int // сколько элементов сейчас
	isFull bool
	mutex  sync.RWMutex
}

// Создать новый буфер
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		events: make([]Event, size),
		size:   size,
		head:   0,
		tail:   0,
		count:  0,
		isFull: false,
	}
}

// Добавить событие
func (cb *CircularBuffer) Add(event Event) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.events[cb.tail] = event

	if cb.isFull {
		// Буфер уже полон, сдвигаем head
		cb.head = (cb.head + 1) % cb.size
	} else {
		// Увеличиваем count пока буфер не полон
		cb.count++
		if cb.count == cb.size {
			cb.isFull = true
		}
	}

	cb.tail = (cb.tail + 1) % cb.size
}

// Получить все события (от старых к новым)
func (cb *CircularBuffer) GetAll() []Event {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.count == 0 {
		return []Event{}
	}

	result := make([]Event, cb.count)

	if !cb.isFull {
		// Буфер не полон, копируем от 0 до tail
		copy(result, cb.events[:cb.tail])
	} else {
		// Буфер полон, копируем от head до конца, потом от начала до tail
		n := copy(result, cb.events[cb.head:])
		copy(result[n:], cb.events[:cb.tail])
	}

	return result
}

// Получить последние N событий
func (cb *CircularBuffer) GetLast(n int) []Event {
	all := cb.GetAll()
	if len(all) <= n {
		return all
	}
	return all[len(all)-n:]
}

// Получить события за последние X минут
func (cb *CircularBuffer) GetRecent(duration time.Duration) []Event {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	var result []Event

	for _, event := range cb.GetAll() {
		if event.Timestamp.After(cutoff) {
			result = append(result, event)
		}
	}

	return result
}

// Timeline - основной модуль
type Timeline struct {
	buffer    *CircularBuffer
	eventChan chan Event
	stopChan  chan struct{}
}

// Создать новый Timeline
func NewTimeline(bufferSize int) *Timeline {
	tl := &Timeline{
		buffer:    NewCircularBuffer(bufferSize),
		eventChan: make(chan Event, 100), // буферизованный канал
		stopChan:  make(chan struct{}),
	}

	// Запускаем горутину для обработки событий
	go tl.processEvents()

	return tl
}

// Добавить событие в канал
func (tl *Timeline) AddEvent(event Event) {

	select {
	case tl.eventChan <- event:
		// logger.Infof("added event:%s", event.Content)
		// Событие добавлено
	default:
		logger.Errorf("cant write event %s: channel is full", event.Content)
		// Канал переполнен, пропускаем (или логируем ошибку)
	}
}

// Горутина для обработки событий из канала
func (tl *Timeline) processEvents() {
	for {
		select {
		case event := <-tl.eventChan:
			tl.buffer.Add(event)
		case <-tl.stopChan:
			return
		}
	}
}

// Методы для получения данных
func (tl *Timeline) GetAllEvents() []Event {
	return tl.buffer.GetAll()
}

func (tl *Timeline) GetLastEvents(n int) []Event {
	return tl.buffer.GetLast(n)
}

func (tl *Timeline) GetRecentEvents(duration time.Duration) []Event {
	return tl.buffer.GetRecent(duration)
}

// Остановить Timeline
func (tl *Timeline) Stop() {
	close(tl.stopChan)
}

func NewEventMock() func() Event {
	i := -1
	return func() Event {
		i++
		return Event{
			Type:      EventChat,
			Content:   fmt.Sprintf("%d", i),
			Author:    fmt.Sprintf("user_%d", i),
			Streamer:  "forsen",
			Timestamp: time.Now(),
		}
	}
}

// TODO: доделать
func SprintEvents(events []Event) string {
	sb := strings.Builder{}
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("%s: %s\n", e.Author, e.Content))

	}
	return sb.String()
}

func PrintEvents(events []Event) {
	sb := strings.Builder{}
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", e.Streamer, e.Author, e.Content))

	}
	fmt.Println(sb.String())
	return
}
