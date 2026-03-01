package formularytest

import "fmt"

// StartStandalone boots an in-process Formulary Service without requiring *testing.T.
// Returns the environment and a cleanup function.
func StartStandalone(tmpDir string) (*Env, func(), error) {
	env, cleanup, err := boot(tmpDir)
	if err != nil {
		return nil, cleanup, fmt.Errorf("formulary: %w", err)
	}
	return env, cleanup, nil
}
