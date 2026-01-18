package task

import (
	"bytes"
	"testing"
)

func TestIO(t *testing.T) {
	data := "o rato roeu a roupa do rei de roma que raivoso rasgou o resto"
	r := bytes.NewBufferString(data)
	w := bytes.NewBuffer([]byte{})
	task := NewCopyTask(r, w, make([]byte, 1), len(data))
	go task.Run()
	for range task.TaskStatus.GetTicker() {
		t.Logf("%d/%d done?:%v - %s\n", task.TaskStatus.Done(), task.TaskStatus.Todo()+task.TaskStatus.Done(), task.TaskStatus.IsCompleted(), task.TaskStatus.Status())
	}
	ret := w.String()
	if data != ret {
		t.Errorf("expected %s got %s", data, ret)
	}
}
