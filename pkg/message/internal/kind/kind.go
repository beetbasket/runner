package kind

import "reflect"

type Kind[K any] struct{}

func (Kind[K]) String() string {
	return reflect.TypeFor[K]().Name()
}
