package command

import (
	"fmt"

	"github.com/ganeshrockz/go-redis/core/resp"
	"github.com/ganeshrockz/go-redis/core/store"
)

func RegisterSetCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     "set",
		Validate: validateSet(),
		Execute:  executeSet(),
	})
}

func validateSet() ValidationHook {
	return func(args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("expected 2 argument, got %d", len(args))
		}

		return nil
	}
}

func executeSet() ExecutionHook {
	return func(args []string, store store.Store) (string, error) {
		err := store.Set(args[0], args[1])
		if err != nil {
			return "", err
		}

		return resp.Encode(resp.RESP_OK), nil
	}
}
