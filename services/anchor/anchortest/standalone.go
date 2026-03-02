package anchortest

import "fmt"

// StartStandalone boots an in-process Anchor Service without requiring *testing.T.
// Returns the environment and a cleanup function.
func StartStandalone(tmpDir string) (*Env, func(), error) {
	env, cleanup, err := boot(tmpDir)
	if err != nil {
		return nil, cleanup, fmt.Errorf("anchor: %w", err)
	}
	return env, cleanup, nil
}
