package container

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"strings"
)

type componentShell struct {
	name      string
	realType  reflect.Type  // underlying type
	realValue reflect.Value // underlying value
	template  interface{}
	fields    map[string]reflect.Value // [dependency name:struct field that requires the dependency]
}

func newComponentShell(name string, realVal reflect.Value, realType reflect.Type, template interface{}) *componentShell {
	return &componentShell{
		name:      name,
		fields:    make(map[string]reflect.Value),
		realType:  realType,
		realValue: realVal,
		template:  template,
	}
}

// DIContainer Dependency injection container, not thread safe
type DIContainer struct {
	depGraph      map[string]*componentShell
	components    map[string]interface{}
	interfaceDeps []reflect.Type
}

// NewDIContainer create new DI container
func NewDIContainer() *DIContainer {
	return &DIContainer{
		depGraph:   make(map[string]*componentShell),
		components: make(map[string]interface{}),
	}
}

var zeroValue = reflect.Value{}

func getQualifiedTypeName(stub interface{}) string {
	var t reflect.Type
	switch stub.(type) {
	case reflect.StructField:
		t = stub.(reflect.StructField).Type
	case reflect.Type:
		t = stub.(reflect.Type)
	case reflect.Value:
		rv := stub.(reflect.Value)
		uv := reflect.Indirect(stub.(reflect.Value))
		if uv.IsValid() {
			t = uv.Type()
		} else {
			// stub is zero value
			t = rv.Type()
		}
	default:
		panic(fmt.Errorf("unsupported stub type '%s', expected reflect.Type, reflect.Type or reflect.Value", t.String()))
	}
	if t.Kind() == reflect.Ptr {
		// if t is pointer type, return its underlying type
		// and if t is the zero value of a type, this call won't panic
		t = t.Elem()
	}
	pkg := t.PkgPath()
	parts := strings.Split(t.String(), ".")
	return path.Join(pkg, parts[len(parts)-1])
}

func getFieldDepName(field reflect.StructField) string {
	v, ok := field.Tag.Lookup("dep")
	if !ok {
		return ""
	}
	if v != "" {
		return v
	}
	return getQualifiedTypeName(field)
}

func isInterfaceType(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Interface
}

// Register register component to DI container by its type name
func (dic *DIContainer) Register(shell interface{}) {
	ptrVal := reflect.ValueOf(shell)
	realVal := reflect.Indirect(ptrVal)
	realType := realVal.Type()
	typeName := getQualifiedTypeName(realType)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error occurred while registering component '%s': %v", typeName, err)
			panic(err)
		}
	}()

	// check availablity
	if ptrVal.Kind() != reflect.Ptr {
		panic(fmt.Errorf("shell parameter must be of pointer type"))
	}
	if realType.Kind() != reflect.Struct {
		panic(fmt.Errorf("component must be of Struct type"))
	}

	cs := newComponentShell(typeName, realVal, realType, shell)
	// make key available, if component has no dependency, entry would not exists
	n := realType.NumField()
	for i := 0; i < n; i++ {
		// field to get tag data
		tField := realType.Field(i)
		// field to set field value
		vField := realVal.Field(i)
		if dep := getFieldDepName(tField); dep != "" {
			if !vField.CanSet() {
				panic(fmt.Errorf("field '%s' should be exported", tField.Name))
			}
			if isInterfaceType(tField) {
				dic.interfaceDeps = append(dic.interfaceDeps, tField.Type)
			}
			cs.fields[dep] = vField
		}
	}
	dic.depGraph[typeName] = cs
}

func callComponentConstructor(template reflect.Value) (instance interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			instance = nil
			err = fmt.Errorf("%v", e)
		}
	}()
	constructor := template.MethodByName("Constructor")
	if constructor == zeroValue {
		return nil, fmt.Errorf("constructor method is missing")
	}
	retVals := constructor.Call(nil)
	if count := len(retVals); count > 1 {
		return nil, fmt.Errorf("too many return values, expect: %d, actual: %d", 1, count)
	}
	return retVals[0].Interface(), nil
}

func initComponent(name string, depGraph map[string]*componentShell, components map[string]interface{},
	pathMap map[string]bool) (interface{}, error) {
	cs := depGraph[name]
	if cs == nil {
		return nil, fmt.Errorf("'%s' is not provided(registered)", name)
	}

	realVal := cs.realValue
	pathMap[name] = true
	for dep, field := range cs.fields {
		var componentPtr interface{}

		if _, ok := components[dep]; !ok {
			if pathMap[dep] { // cycle detected
				return nil, fmt.Errorf("cycle dependency detected, '%s' and '%s' are depend on each other", name, dep)
			}
			ptr, err := initComponent(dep, depGraph, components, pathMap)
			if err != nil {
				return nil, err
			}
			components[dep] = ptr
			componentPtr = ptr
		}
		if depVal := reflect.ValueOf(componentPtr); !depVal.Type().AssignableTo(field.Type()) {
			return nil, fmt.Errorf("'%s' is not assignable to '%s'", getQualifiedTypeName(depVal), getQualifiedTypeName(field))
		}
		field.Set(reflect.ValueOf(componentPtr))
	}
	// Constructor is value type receiver's to call
	componentPtr, err := callComponentConstructor(realVal)
	if err != nil {
		return nil, fmt.Errorf("failed to call Constructor of '%s': %v", name, err)
	}
	components[name] = componentPtr
	pathMap[name] = false
	return componentPtr, nil
}

// Get return component from DI container by component name, initialization may be needed
func (dic *DIContainer) Get(name string) interface{} {
	components := dic.components
	if c, ok := components[name]; ok {
		return c
	}
	// record init path
	pathMap := make(map[string]bool)
	pathMap[name] = true
	c, err := initComponent(name, dic.depGraph, dic.components, pathMap)
	if err != nil {
		log.Printf("Error occurred while injecting dependency of '%s':\n  %v", name, err)
		panic(err)
	}
	return c
}
