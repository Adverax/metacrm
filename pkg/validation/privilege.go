package validation

import (
	"context"
)

var (
	// ErrMissingPrivilege is the error returned when a user does not have the required privilege.
	ErrMissingPrivilege = NewError("validation_missing_privilege", "user does not have the required privilege")
)

type PrivilegeResolver interface {
	HasPrivilege(ctx context.Context, resource, action string) (bool, error)
}

// Privilege is a constructor for a privilege validation rule.
func Privilege(
	resource string,
	action string,
) PrivilegeRule {
	return PrivilegeRule{
		condition: true,
		resource:  resource,
		action:    action,
		err:       ErrMissingPrivilege,
	}
}

// PrivilegeRule checks if the user has the required privilege.
type PrivilegeRule struct {
	condition bool
	resource  string
	action    string
	err       Error
}

func (that PrivilegeRule) Validate(ctx context.Context, value interface{}) error {
	if !that.condition {
		return nil
	}

	resolver := GetPrivilegeResolver(ctx)
	if resolver == nil {
		return nil
	}

	granted, err := resolver.HasPrivilege(ctx, that.resource, that.action)
	if err != nil {
		return err
	}

	if !granted {
		return ErrMissingPrivilege.SetParams(map[string]interface{}{"action": that.action})
	}

	return nil
}

// Error sets the error message for the rule.
func (that PrivilegeRule) Error(message string) PrivilegeRule {
	if that.err == nil {
		that.err = ErrNotNilRequired
	}
	that.err = that.err.SetMessage(message)
	return that
}

// ErrorObject sets the error struct for the rule.
func (that PrivilegeRule) ErrorObject(err Error) PrivilegeRule {
	that.err = err
	return that
}

// When sets the condition that determines if the validation should be performed.
func (that PrivilegeRule) When(condition bool) PrivilegeRule {
	that.condition = condition
	return that
}

type privilegeCtxType int

const privilegeCtxKey privilegeCtxType = 1

// GetPrivilegeResolver retrieves the PrivilegeResolver from the context.
func GetPrivilegeResolver(ctx context.Context) PrivilegeResolver {
	if ctx == nil {
		return nil
	}

	resolver, ok := ctx.Value(privilegeCtxKey).(PrivilegeResolver)
	if !ok {
		return nil
	}
	return resolver
}

// WithPrivilegeResolver adds a PrivilegeResolver to the context.
func WithPrivilegeResolver(ctx context.Context, resolver PrivilegeResolver) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, privilegeCtxKey, resolver)
}
