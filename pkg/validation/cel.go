package validation

import (
	"context"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types/ref"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type Declarations interface {
	DeclareInteger(key string, value int)
	DeclareFloat(key string, value float64)
	DeclareString(key string, value string)
	DeclareBool(key string, value bool)
	DeclareTime(key string, value time.Time)
}

type Declarator interface {
	DeclareValidationFields(declarations Declarations)
}

// ErrCELEnvironment is the error that returns when CEL environment cannot be created.
var ErrCELEnvironment = NewError("validation_cel_error", "can't create CEL environment")

// ErrCELCompilation is the error that returns when CEL compilation fails.
var ErrCELCompilation = NewError("validation_cel_compilation_error", "CEL compilation error")

// ErrCELProgram is the error that returns when CEL program cannot be created.
var ErrCELProgram = NewError("validation_cel_program_error", "error creating CEL program")

// ErrCELEvaluation is the error that returns when CEL evaluation fails.
var ErrCELEvaluation = NewError("validation_cel_evaluation_error", "error evaluating CEL expression")

type celErrors struct {
	errEnv  Error // Error when CEL environment cannot be created
	errComp Error // Error when CEL compilation fails
	errProg Error // Error when CEL program cannot be created
	errEval Error // Error when CEL evaluation fails
}

var defaultCELErrors = celErrors{
	errEnv:  ErrCELEnvironment,
	errComp: ErrCELCompilation,
	errProg: ErrCELProgram,
	errEval: ErrCELEvaluation,
}

// validateCriteria checks if the given value is valid or not.
func validateCriteria(
	ctx context.Context,
	expression string,
	value interface{},
	errs *celErrors,
) (bool, error) {
	declarations, input := declareFields(ctx, value)
	env, err := newEnv(declarations)
	if err != nil {
		return false, errs.errEnv.SetParams(map[string]interface{}{"error": err.Error()})
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return false, errs.errEnv.SetParams(map[string]interface{}{"error": issues.Err().Error()})
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, errs.errProg.SetParams(map[string]interface{}{"error": err.Error()})
	}

	out, _, err := prg.Eval(input)
	if err != nil {
		return false, errs.errEval.SetParams(map[string]interface{}{"error": err.Error()})
	}

	return checkCelResult(out), nil
}

func declareFields(ctx context.Context, this any) ([]*exprpb.Decl, map[string]any) {
	declarations := fieldDeclarationsImpl{
		input: map[string]any{
			"this": this,
		},
	}

	declarator := getFieldDeclarator(ctx)
	if declarator != nil {
		declarator.DeclareValidationFields(&declarations)
	}

	return declarations.decls, declarations.input
}

func newEnv(declarations []*exprpb.Decl) (*cel.Env, error) {
	return cel.NewEnv(
		cel.Declarations(
			append(declarations, decls.NewVar("this", decls.Dyn))...,
		),
	)
}

func checkCelResult(val ref.Val) bool {
	switch v := val.Value().(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int64:
		return v != 0
	case uint64:
		return v != 0
	case float64:
		return v != 0.0
	default:
		return false
	}
}

type fieldDeclarationsImpl struct {
	decls []*exprpb.Decl
	input map[string]any
}

func (that *fieldDeclarationsImpl) DeclareInteger(key string, value int) {
	that.decls = append(that.decls, decls.NewVar(key, decls.Int))
	that.input[key] = value
}

func (that *fieldDeclarationsImpl) DeclareFloat(key string, value float64) {
	that.decls = append(that.decls, decls.NewVar(key, decls.Double))
	that.input[key] = value
}

func (that *fieldDeclarationsImpl) DeclareString(key string, value string) {
	that.decls = append(that.decls, decls.NewVar(key, decls.String))
	that.input[key] = value
}

func (that *fieldDeclarationsImpl) DeclareBool(key string, value bool) {
	that.decls = append(that.decls, decls.NewVar(key, decls.Bool))
	that.input[key] = value
}

func (that *fieldDeclarationsImpl) DeclareTime(key string, value time.Time) {
	that.decls = append(that.decls, decls.NewVar(key, decls.Timestamp))
	that.input[key] = value
}

// getFieldDeclarator retrieves the Declarator from the context.
func getFieldDeclarator(ctx context.Context) Declarator {
	this := GetThis(ctx)
	if this == nil {
		return nil
	}

	if declarator, ok := this.(Declarator); ok {
		return declarator
	}

	return nil
}
