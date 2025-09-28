package access

import "context"

type dummyGetterSetter struct{}

func (d *dummyGetterSetter) GetProperty(
	ctx context.Context,
	name string,
) (interface{}, error) {
	return nil, nil
}

func (d *dummyGetterSetter) SetProperty(
	ctx context.Context,
	name string,
	value interface{},
) error {
	return nil
}

var aDummyGetterSetter = &dummyGetterSetter{}

// NewDummyGetterSetter is simgleton constructor for build dummy GetterSetter
func NewDummyGetterSetter() GetterSetter {
	return aDummyGetterSetter
}
