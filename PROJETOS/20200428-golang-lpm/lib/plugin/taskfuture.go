package plugin

type TaskFuture struct {
	DoneChan     chan error
	taskFunction func(TaskStatusChanger) error
	TaskStatus   TaskStatus
	isStarted    bool
}

func NewTaskFuture(f func(TaskStatusChanger) error) TaskFuture {
	return TaskFuture{
		taskFunction: f,
		DoneChan:     make(chan error),
		TaskStatus:   NewTaskStatus(),
	}
}

func (tf *TaskFuture) Run() error {
	if tf.isStarted {
		return <-tf.DoneChan
	}
	tf.isStarted = true
	err := tf.taskFunction(tf.TaskStatus.(TaskStatusChanger))
sendfeedback:
	select {
	case tf.DoneChan <- err:
		goto sendfeedback
	default:
		defer close(tf.DoneChan)
	}
	tf.TaskStatus.(TaskStatusChanger).SetIsCompleted(true)
	defer close(tf.TaskStatus.GetTicker())
	return err
}
