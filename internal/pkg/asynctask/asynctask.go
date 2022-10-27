package asynctask

import (
	"sync"
	"time"
)

type Manager struct {
	tasks    map[string]*Task
	listener Listener
	l        sync.Mutex
}

func NewManager() *Manager {
	return &Manager{tasks: map[string]*Task{}}
}

type Event struct {
	IsDone bool
	Log    string
}

type Listener func(task *Task, event *Event)

type Task struct {
	Key       string
	m         *Manager
	Logs      []string
	createdAt time.Time
}

func (t *Task) Write(p []byte) (n int, err error) {
	t.Log(string(p))
	return len(p), nil
}

func (t *Task) Log(l string) {
	t.Logs = append(t.Logs, l)
	t.m.Emit(t, &Event{
		IsDone: false,
		Log:    l,
	})
}

func (t *Task) Done() {
	t.m.Emit(t, &Event{
		IsDone: true,
		Log:    "",
	})

	t.m.RemoveTask(t)
}

func (m *Manager) Emit(t *Task, event *Event) {
	if m.listener != nil {
		m.listener(t, event)
	}
}

func (m *Manager) RemoveTask(t *Task) {
	m.l.Lock()
	defer m.l.Unlock()
	delete(m.tasks, t.Key)
}

func (m *Manager) NewTask(key string) (t *Task, isNew bool) {
	m.l.Lock()
	defer m.l.Unlock()

	if t, ok := m.tasks[key]; ok {
		return t, false
	}

	t = &Task{
		Key:       key,
		m:         m,
		Logs:      nil,
		createdAt: time.Now(),
	}
	m.tasks[key] = t
	isNew = true
	return
}

func (m *Manager) AddListener(l Listener) {
	m.listener = l
	return
}
