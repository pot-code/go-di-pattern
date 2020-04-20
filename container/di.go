package container

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type componentTemplate struct {
	instance interface{}
	fields   map[string]reflect.Value // [dependency name:struct field that requires the dependency]
}

// DIContainer Dependency injection container, not thread safe
type DIContainer struct {
	depGraph   map[string][]string
	components map[string]interface{}
	templates  map[string]*componentTemplate
}

// NewDIContainer create new DI container
func NewDIContainer() *DIContainer {
	return &DIContainer{
		make(map[string][]string),
		make(map[string]interface{}),
		make(map[string]*componentTemplate)}
}

var zeroValue = reflect.Value{}

// Register register component to DI container by its type name
func (dic *DIContainer) Register(shell interface{}, namespace ...string) {
	shallowVal := reflect.ValueOf(shell)
	val := reflect.Indirect(shallowVal)
	valType := val.Type()
	typeName := valType.Name()

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error occurred while registering component '%s': %v", typeName, err)
			panic(err)
		}
	}()

	// check availablity
	if shallowVal.Kind() != reflect.Ptr {
		panic(fmt.Errorf("shell parameter must be of pointer type"))
	}
	if valType.Kind() != reflect.Struct {
		panic(fmt.Errorf("component must be of Struct type"))
	}

	if len(namespace) > 0 {
		typeName = strings.Join(namespace, ":") + ":" + typeName
	}
	n := valType.NumField()
	ct := &componentTemplate{instance: shell, fields: make(map[string]reflect.Value)}
	// make key available, if component has no dependency, entry would not exists
	dic.depGraph[typeName] = nil
	for i := 0; i < n; i++ {
		// field to get tag data
		tField := valType.Field(i)
		// field to set field value
		vField := val.Field(i)
		if dep := tField.Tag.Get("dep"); dep != "" {
			if !vField.CanSet() {
				panic(fmt.Errorf("field '%s' should be exported", tField.Name))
			}
			dic.depGraph[typeName] = append(dic.depGraph[typeName], dep)
			ct.fields[dep] = vField
		}
	}
	dic.templates[typeName] = ct
}

func callComponentConstructor(shell reflect.Value) (instance interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			instance = nil
			err = fmt.Errorf("%v", e)
		}
	}()
	constructor := shell.MethodByName("Constructor")
	if constructor == zeroValue {
		return nil, fmt.Errorf("constructor method is missing")
	}
	retVals := constructor.Call(nil)
	if count := len(retVals); count > 1 {
		return nil, fmt.Errorf("too many return values, expect: %d, actual: %d", 1, count)
	}
	return retVals[0].Interface(), nil
}

func initComponent(name string, depGraph map[string][]string, components map[string]interface{},
	templates map[string]*componentTemplate, pathMap map[string]bool) (interface{}, error) {
	deps := depGraph[name]
	ci := templates[name]

	if ci == nil {
		return nil, fmt.Errorf("'%s' is not provided(registered)", name)
	}

	val := reflect.Indirect(reflect.ValueOf(ci.instance))
	pathMap[name] = true
	for _, dep := range deps {
		if c, ok := components[dep]; ok {
			ci.fields[dep].Set(reflect.ValueOf(c))
		} else {
			if pathMap[dep] { // cycle detected
				return nil, fmt.Errorf("cycle dependency detected, '%s' and '%s' are depend on each other", name, dep)
			}
			depComponentPtr, err := initComponent(dep, depGraph, components, templates, pathMap)
			if err != nil {
				return nil, err
			}
			components[dep] = depComponentPtr
			ci.fields[dep].Set(reflect.ValueOf(depComponentPtr))
		}
	}

	componentPtr, err := callComponentConstructor(val)
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
	c, err := initComponent(name, dic.depGraph, dic.components, dic.templates, pathMap)
	if err != nil {
		log.Printf("Error occurred while injecting dependency of '%s':\n  %v", name, err)
		panic(err)
	}
	return c
}
