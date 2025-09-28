package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type dummyErrorBuilder struct{}

func (that *dummyErrorBuilder) Build(_ context.Context, err error, _ string) error {
	return err
}

// FreeContext - Free context from scope
func FreeContext(ctx context.Context, key DbId) context.Context {
	if ctx == nil {
		return nil
	}
	if key == "" {
		return ctx
	}
	val := ctx.Value(key)
	if _, valid := val.(Scope); valid {
		return context.WithValue(ctx, key, nil)
	}
	return ctx
}

// FromContext - Extract scope from context
func FromContext(ctx context.Context, key DbId) Scope {
	val := ctx.Value(key)
	if c, valid := val.(Scope); valid {
		return c
	}
	return nil
}

// ToContext - Append scope into context
func ToContext(
	ctx context.Context,
	scope Scope,
) context.Context {
	return context.WithValue(ctx, scope.DbId(ctx), scope)
}

// FetchModel - returns a single model from fetcher
func FetchModel[T any](
	fetcher Fetcher,
	reader func(scanner Scanner) (T, error),
) (model T, err error) {
	err = fetcher(func(rows Rows) error {
		var err error
		model, err = reader(rows)
		if err != nil {
			if errors.Is(err, ErrSkipValue) {
				return ErrNoRows
			}
			return fmt.Errorf("read query: %w", err)
		}
		return nil
	})
	if err != nil {
		return model, err
	}

	return model, nil
}

// FetchModels - returns multiple models from fetcher
func FetchModels[T any](
	fetcher Fetcher,
	reader func(scanner Scanner) (T, error),
) (models []T, err error) {
	err = fetcher(func(rows Rows) error {
		val, err := reader(rows)
		if err != nil {
			if errors.Is(err, ErrSkipValue) {
				return nil // Skip this value
			}
			return fmt.Errorf("read query: %w", err)
		}
		models = append(models, val)
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNoRows) {
			return nil, nil // No rows found
		}
		return nil, err
	}
	return models, nil
}

// FetchModelMap - returns multiple models from fetcher as a map
func FetchModelMap[K comparable, T any](
	fetcher Fetcher,
	reader func(scanner Scanner) (K, T, error),
) (models map[K]T, err error) {
	err = fetcher(func(rows Rows) error {
		models = make(map[K]T)
		key, val, err := reader(rows)
		if err != nil {
			if errors.Is(err, ErrSkipValue) {
				return nil // Skip this value
			}
			return fmt.Errorf("read query: %w", err)
		}
		models[key] = val
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNoRows) {
			return nil, nil // No rows found
		}
		return nil, err
	}
	return models, nil
}

// FetchDictionary - returns a dictionary from fetcher
func FetchDictionary[T comparable, V any](
	fetcher Fetcher,
) (res map[T]V, err error) {
	res = make(map[T]V, 256)
	err = fetcher(func(rows Rows) error {
		var key T
		var val V
		err := rows.Scan(&key, &val)
		if err != nil {
			if errors.Is(err, ErrSkipValue) {
				return nil // Skip this value
			}
			return fmt.Errorf("read query: %w", err)
		}
		res[key] = val
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNoRows) {
			return nil, nil // No rows found
		}
		return nil, err
	}
	return res, nil
}

// FetchMap - returns a map from fetcher
func FetchMap[K comparable](fetcher Fetcher) (map[K]any, error) {
	var result map[K]any
	err := fetcher(func(rows Rows) error {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}
		result, err = scanMap[K](rows, columns)
		if err != nil {
			return err
		}
		return nil
	})
	if err == nil && result == nil {
		return nil, ErrNoRows
	}
	return result, err
}

// FetchMaps - returns a slice of maps from fetcher
func FetchMaps[K comparable](fetcher Fetcher) ([]map[K]any, error) {
	result := make([]map[K]any, 0)
	err := fetcher(func(rows Rows) error {
		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}
		m, err := scanMap[K](rows, columns)
		if err != nil {
			return fmt.Errorf("failed to scan properties: %w", err)
		}
		result = append(result, m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanMap[K comparable](reader Scanner, columns []string) (map[K]any, error) {
	values := make([]any, len(columns))
	vs := make([]any, len(columns))
	for i := range values {
		vs[i] = &values[i]
	}
	if err := reader.Scan(vs...); err != nil {
		return nil, err
	}
	props := make(map[K]any)
	for i, col := range columns {
		k := reflect.ValueOf(col).Convert(reflect.TypeOf((*K)(nil)).Elem()).Interface().(K)
		props[k] = values[i]
	}
	return props, nil
}

func newFetcher(ctx context.Context, scope Scope, query string, args ...any) Fetcher {
	return func(scan func(rows Rows) error) error {
		rs, err := scope.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("Query: %w", err)
		}
		defer rs.Close()

		for rs.Next() {
			if err := scan(rs); err != nil {
				return fmt.Errorf("read query: %w", err)
			}
		}

		if err := rs.Err(); err != nil {
			if !errors.Is(err, ErrNoRows) {
				return err
			}
		}

		return nil
	}
}

var ErrSkipValue = errors.New("skip value")

func MakeLiteral(v any) string {
	if v == nil {
		return "NULL"
	}

	switch x := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(x, "'", "''") + "'"
	case int, int64:
		return fmt.Sprintf("%v", x)
	case float64:
		// Используем strconv.FormatFloat с автоматической точностью
		// 'g' формат автоматически выбирает между 'f' и 'e' для оптимального представления
		return strconv.FormatFloat(x, 'g', -1, 64)
	case bool:
		return fmt.Sprintf("%v", x)
	default:
		b, _ := json.Marshal(x)
		return "'" + string(b) + "'"
	}
}
