package main

import (
	"errors"
	"sync"

	"github.com/BSFishy/mora-manager/state"
)

type FunctionContext struct {
	Registry   *FunctionRegistry
	Config     *Config
	State      *state.State
	ModuleName string
}

type ExpressionFunction struct {
	MinArgs         int
	MaxArgs         int // -1 for unlimited
	Evaluate        func(FunctionContext, Args) (*ReturnType, error)
	GetConfigPoints func(FunctionContext, Args) ([]ConfigPoint, error)
}

func (e *ExpressionFunction) InvalidArgs(args Args) bool {
	len := args.Len()
	if e.MinArgs > len {
		return true
	}

	if e.MaxArgs != -1 && e.MaxArgs < len {
		return true
	}

	return false
}

type TypeEnum int

const (
	Identifier TypeEnum = iota
	String
)

type ReturnType struct {
	Type  TypeEnum
	Value any
}

func NewIdentifier(value string) ReturnType {
	return ReturnType{
		Type:  Identifier,
		Value: value,
	}
}

func NewString(value string) ReturnType {
	return ReturnType{
		Type:  String,
		Value: value,
	}
}

func (r *ReturnType) Identifier() *string {
	if r.Type != Identifier {
		return nil
	}

	value, ok := r.Value.(string)
	if !ok {
		panic("identifier return type is not a string")
	}

	return &value
}

func (r *ReturnType) String() *string {
	if r.Type != String {
		return nil
	}

	value, ok := r.Value.(string)
	if !ok {
		panic("string return type is not a string")
	}

	return &value
}

type FunctionRegistry struct {
	mu      sync.RWMutex
	funcMap map[string]ExpressionFunction
}

func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		funcMap: map[string]ExpressionFunction{},
	}
}

func (r *FunctionRegistry) Register(name string, fn ExpressionFunction) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.funcMap[name] = fn
}

func (r *FunctionRegistry) Get(name string) (ExpressionFunction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, ok := r.funcMap[name]
	return fn, ok
}

type Args []Expression

func (a Args) Len() int {
	return len(a)
}

func (a Args) Identifier(ctx FunctionContext, i int) (string, error) {
	if i >= len(a) {
		return "", errors.New("argument index out of range")
	}

	expr := a[i]
	return expr.EvaluateIdentifier(ctx)
}

func (a Args) String(ctx FunctionContext, i int) (string, error) {
	if i >= len(a) {
		return "", errors.New("argument index out of range")
	}

	expr := a[i]
	return expr.EvaluateString(ctx)
}
