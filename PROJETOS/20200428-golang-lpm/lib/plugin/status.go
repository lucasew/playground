package plugin

type TaskStatus interface {
	Done() uint
	Status() string
	Todo() uint
	IsCompleted() bool
	GetTicker() chan struct{}
}

type TaskStatusChanger interface {
	SetDone(uint)
	SetStatus(string)
	SetTodo(uint)
	SetIsCompleted(bool)
}

type taskStatus struct {
	done        uint
	todo        uint
	isCompleted bool
	status      string
	tick        chan struct{}
}

func NewTaskStatus() *taskStatus {
	return &taskStatus{
		status: "initializing...",
		tick:   make(chan struct{}, 1),
	}
}

func (ts *taskStatus) SetDone(v uint) {
	defer ts.tickChange()
	ts.done = v
}

func (ts *taskStatus) SetTodo(v uint) {
	defer ts.tickChange()
	ts.todo = v
}

func (ts *taskStatus) SetIsCompleted(v bool) {
	defer ts.tickChange()
	ts.isCompleted = v
}
func (ts *taskStatus) SetStatus(v string) {
	defer ts.tickChange()
	ts.status = v
}

func (ts taskStatus) tickChange() {
	select {
	case ts.tick <- struct{}{}:
	default:
		return
	}
}

func (ts taskStatus) Done() uint {
	return ts.done
}

func (ts taskStatus) Todo() uint {
	return ts.todo
}

func (ts taskStatus) Status() string {
	return ts.status
}

func (ts taskStatus) IsCompleted() bool {
	return ts.isCompleted
}

func (ts taskStatus) GetTicker() chan struct{} {
	return ts.tick
}

func init() {
	_ = TaskStatus(NewTaskStatus())
	_ = TaskStatusChanger(NewTaskStatus())
}
