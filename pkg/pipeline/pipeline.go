package pipeline

// A pipeline will have stages where the output of each subsequent stage will be the input of the following stage
// Sn...(S3(S2(S1(i))))
// We can use channels to pass data from one stage to the other

type (
	Stage[T any]          func() chan T
	StageProcessor[T any] func(Stage[T]) Stage[T]
)

func Pipe[T any](s Stage[T], sps ...StageProcessor[T]) Stage[T] {
	for _, process := range sps {
		s = process(s)
	}
	return s
}
