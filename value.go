package gsp

type Value[X any] interface {
	Jobs() []Job[X]
}
