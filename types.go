package toolkit

import "reflect"

func GetType(instance any) string {
	return reflect.TypeOf(instance).String()
}
