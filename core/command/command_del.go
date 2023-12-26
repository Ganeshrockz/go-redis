package command

import (
	"fmt"

	"github.com/ganeshrockz/go-redis/core/resp"
	"github.com/ganeshrockz/go-redis/core/store"
)

func RegisterDeleteCommand(r CommandRegistry) {
	r.Add(&CommandRegistration{
		Name:     "del",
		Validate: validateDel(),
		Execute:  executeDel(),
	})
}

func validateDel() ValidationHook {
	return func(args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected 1 argument, got %d", len(args))
		}

		return nil
	}
}

func executeDel() ExecutionHook {
	return func(args []string, store store.Store) (string, error) {
		err := store.Delete(args[0])
		if err != nil {
			return "", err
		}

		return resp.Encode(resp.RESP_OK), nil
	}
}
