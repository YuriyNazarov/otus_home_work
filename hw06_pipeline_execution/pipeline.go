package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	outChan := make(Bi)
	if len(stages) == 0 {
		close(outChan)
		return outChan
	}

	prevStageOut := in
	for _, stage := range stages {
		prevStageOut = stage(prevStageOut)
	}

	go func() {
		defer close(outChan)
		for {
			select {
			case <-done:
				return
			case buf, ok := <-prevStageOut:
				if ok {
					outChan <- buf
				} else {
					return
				}
			}
		}
	}()
	return outChan
}
