package container

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type componentTemplate struct {
	instance interface{}
	fields   map[string]reflect.Value
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

// Register register component to DI container by its type name
func (dic *DIContainer) Register(shell interface{}, namespace ...string) {
	val := reflect.Indirect(reflect.ValueOf(shell))
	tp := val.Type()
	typeName := tp.Name()

	if len(namespace) > 0 {
		typeName = strings.Join(namespace, ":") + ":" + typeName
	}
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("Instance of %s is not a type of struct", typeName)
		}
	}()

	n := tp.NumField()
	ct := &componentTemplate{instance: shell, fields: make(map[string]reflect.Value)}
	// make key available, if component has no dependency, entry would not exists
	dic.depGraph[typeName] = nil
	for i := 0; i < n; i++ {
		// field to get tag data
		tField := tp.Field(i)
		// field to set field value
		vField := val.Field(i)
		if dep := tField.Tag.Get("dep"); dep != "" {
			dic.depGraph[typeName] = append(dic.depGraph[typeName], dep)
			ct.fields[dep] = vField
		}
	}
	dic.templates[typeName] = ct
}

func initComponent(name string, depGraph map[string][]string, components map[string]interface{},
	templates map[string]*componentTemplate, pathMap map[string]bool) interface{} {
	deps := depGraph[name]
	ci := templates[name]

	if ci == nil {
		panic(fmt.Errorf("%s not exists in dependency tree", name))
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error occurred while injecting dependency of %s: %v", name, err)
			panic(err)
		}
	}()
	val := reflect.Indirect(reflect.ValueOf(ci.instance))
	pathMap[name] = true
	for _, dep := range deps {
		if c, ok := components[dep]; ok {
			ci.fields[dep].Set(reflect.ValueOf(c))
		} else {
			if pathMap[dep] { // cycle detected
				panic(fmt.Errorf("Cycle dependency detected, %s and %s are depend on each other", name, dep))
			}
			depComponentPtr := initComponent(dep, depGraph, components, templates, pathMap)
			components[dep] = depComponentPtr
			ci.fields[dep].Set(reflect.ValueOf(depComponentPtr))
		}
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Cannot call Constructor of %s: %v", name, err)
			panic(err)
		}
	}()
	componentPtr := val.MethodByName("Constructor").Call(nil)[0].Interface()
	components[name] = componentPtr
	pathMap[name] = false
	return componentPtr
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
	return initComponent(name, dic.depGraph, dic.components, dic.templates, pathMap)
}
