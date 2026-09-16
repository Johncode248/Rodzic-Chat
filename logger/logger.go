package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// redactedSensitiveKeyFragments matches JSON object keys (case-insensitive
// substring) whose value must never reach the logs -- passwords are logged
// in plaintext on every login/register call otherwise, since Enter/Exit
// JSON-marshal whatever request/argument struct they're handed wholesale
// (see getDataFlow below).
var redactedSensitiveKeyFragments = []string{"password"}

const redactedPlaceholder = "[REDACTED]"

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, fragment := range redactedSensitiveKeyFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

// redactSensitive walks a generic json.Unmarshal-produced value (maps/
// slices/scalars) and blanks out any object value whose key looks
// sensitive, in place.
func redactSensitive(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, sub := range val {
			if isSensitiveKey(k) {
				val[k] = redactedPlaceholder
				continue
			}
			redactSensitive(sub)
		}
	case []interface{}:
		for _, sub := range val {
			redactSensitive(sub)
		}
	}
}

var tracer = otel.Tracer("my-service-logger")

func Enter(ctx context.Context, name string, input ...interface{}) context.Context {
	ctx, span := tracer.Start(ctx, name)

	depth := getDepth(ctx)
	processedInput := getDataFlow(input...)

	log.Printf("%s>> %s | TraceID: %s | input: %v", strings.Repeat("  ", depth), name, span.SpanContext().TraceID(), processedInput)

	return context.WithValue(ctx, "depth", depth+1)
}

func Exit(ctx context.Context, name string, output ...interface{}) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	depth := getDepth(ctx) - 1
	if depth < 0 {
		depth = 0
	}

	// Przetwarzamy na JSONy
	processedOutput := getDataFlow(output...)

	hasError := false
	for i, val := range output {
		if isError(val) {
			hasError = true
			// Pobieramy ładny komunikat błędu z przetworzonej listy
			errorMsg := processedOutput[i]
			span.RecordError(fmt.Errorf(errorMsg))
			span.SetStatus(codes.Error, errorMsg)
		}
	}

	statusIcon := "✅"
	if hasError {
		statusIcon = "❌"
	}

	// Wyświetlamy jako stringi JSON połączone przecinkiem
	log.Printf(
		"%s<< %s %s | output: [%s]",
		strings.Repeat("  ", depth),
		name,
		statusIcon,
		strings.Join(processedOutput, ", "),
	)
}

func getDepth(ctx context.Context) int {
	if d, ok := ctx.Value("depth").(int); ok {
		return d
	}
	return 0
}

func isError(val interface{}) bool {
	v := reflect.ValueOf(val)

	// Jeśli to wskaźnik (np. &err z Twojego defer), musimy go "odpakować"
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false // Wskaźnik sam w sobie jest nil
		}
		v = v.Elem() // Dobieramy się do zawartości zmiennej err
	}

	// Sprawdzamy czy to co zostało implementuje interfejs błędu
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !v.Type().Implements(errorType) {
		return false
	}

	// NAJWAŻNIEJSZE: sprawdzamy czy wartość błędu nie jest nil
	// W Go interfejs błędu może istnieć, ale nie nieść żadnego błędu
	return !v.IsNil()
}

func getDataFlow(args ...interface{}) []string { // Zmieniamy powrót na []string dla czytelności
	var result []string
	for _, arg := range args {
		v := reflect.ValueOf(arg)

		// 1. Dereferencja wskaźnika
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				result = append(result, "null")
				continue
			}
			v = v.Elem()
		}

		// 2. Obsługa nil
		if !v.IsValid() {
			result = append(result, "null")
			continue
		}

		// 3. Obsługa Errora
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if v.Type().Implements(errorType) {
			if v.IsNil() {
				result = append(result, "null")
			} else {
				result = append(result, fmt.Sprintf("ERROR: %s", v.Interface().(error).Error()))
			}
			continue
		}

		// 4. Zamiana wszystkiego innego na JSON
		// v.Interface() pobiera wartość, którą json.Marshal może przetworzyć
		b, err := json.Marshal(v.Interface())
		if err != nil {
			// Jeśli JSON zawiedzie, użyj fmt z nazwami pól (%+v)
			result = append(result, fmt.Sprintf("%+v", v.Interface()))
			continue
		}

		var generic interface{}
		if err := json.Unmarshal(b, &generic); err == nil {
			redactSensitive(generic)
			if redacted, err := json.Marshal(generic); err == nil {
				b = redacted
			}
		}

		result = append(result, string(b))
	}
	return result
}
