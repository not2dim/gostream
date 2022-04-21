package stream

import (
	"reflect"
)

func reflectBaseMeta(stm any) (ret *meta) {
	var ok bool
	defer func() {
		if !ok {
			panic("failed to reflect base.Meta")
		}
	}()
	rv := reflect.ValueOf(stm)
	var v2Meta = func(t reflect.Type, v reflect.Value) (ret *meta, ok bool) {
		if t.Kind() == reflect.Pointer {
			v = v.Elem()
			t = v.Type()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			fv := v.Field(i)
			if t.Field(i).Name == "Meta" {
				if fv.Interface() == nil {
					return nil, true
				}
				return fv.Interface().(*meta), true
			}
		}
		return
	}
	ret, ok = v2Meta(rv.Type(), rv)
	if ok {
		return
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		ret, ok = v2Meta(fv.Type(), fv)
		if ok {
			return
		}
	}
	return
}

func reflectBaseCurr(stm any) (ret pipeline) {
	var ok bool
	defer func() {
		if !ok {
			panic("failed to reflect base.Curr")
		}
	}()
	rv := reflect.ValueOf(stm)
	var v2Curr = func(t reflect.Type, v reflect.Value) (ret pipeline, ok bool) {
		if t.Kind() == reflect.Pointer {
			v = v.Elem()
			t = v.Type()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			fv := v.Field(i)
			if t.Field(i).Name == "Curr" {
				if fv.Interface() == nil {
					return nil, true
				}
				return fv.Interface().(pipeline), true
			}
		}
		return
	}
	ret, ok = v2Curr(rv.Type(), rv)
	if ok {
		return
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		ret, ok = v2Curr(fv.Type(), fv)
		if ok {
			return
		}
	}
	return
}

func reflectMapper(mapper any) (mtd reflect.Value) {
	var ok bool
	defer func() {
		if !ok {
			panic("failed to reflect mapper func [S any, T](v S) T")
		}
	}()
	mtd = reflect.ValueOf(mapper)
	if mtd.Kind() != reflect.Func {
		return
	}
	if mtd.Type().NumIn() != 1 || mtd.Type().NumOut() != 1 {
		return
	}
	ok = true
	return mtd
}

func reflectCallMapper(mapper reflect.Value, v any) any {
	ret := mapper.Call([]reflect.Value{reflect.ValueOf(v)})
	if len(ret) != 1 {
		panic("failed to call mapper reflect.Value, also as func [S any, T](v S) T.")
	}
	return ret[0].Interface()
}
