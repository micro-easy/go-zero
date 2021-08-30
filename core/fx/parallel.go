package fx

import "github.com/micro-easy/go-zero/core/threading"

func Parallel(fns ...func()) {
	group := threading.NewRoutineGroup()
	for _, fn := range fns {
		group.RunSafe(fn)
	}
	group.Wait()
}
