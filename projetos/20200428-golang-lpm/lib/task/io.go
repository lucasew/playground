package task

import (
	"io"

	"github.com/lucasew/lpm/lib/plugin"
)

func NewCopyTask(r io.Reader, w io.Writer, buf []byte, sizehint int) plugin.TaskFuture {
	return plugin.NewTaskFuture(func(tc plugin.TaskStatusChanger) error {
		tc.SetStatus("copying data...")
		if buf == nil {
			_, err := io.Copy(w, r)
			return err
		}
		n := 0
		for true {
			size, err := r.Read(buf)
			if io.EOF == err {
				return nil
			}
			if err != nil {
				return err
			}
			n += size
			_, err = w.Write(buf[:size])
			if err != nil {
				return err
			}
			tc.SetDone(uint(n))
			if sizehint >= 0 {
				tc.SetTodo(uint(sizehint - n))
			}
		}
		return nil
	})
}
