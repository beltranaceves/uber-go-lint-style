package varscope

import "os"

func BadSimple() error {
	err := os.WriteFile("f", nil, 0644) // want "identifier 'err' can be declared in the inner block to reduce its scope"
	if err != nil {
		return err
	}
	return nil
}

func GoodShortIf() error {
	if err := os.WriteFile("f", nil, 0644); err != nil {
		return err
	}
	return nil
}

func NoReduceWhenUsedAfter() error {
	err := os.WriteFile("f", nil, 0644)
	if err != nil {
		// do something
	}
	_ = err // used after the if -> should NOT trigger
	return nil
}

func UsedOnlyInsideIf() error {
	x := os.TempDir()
	if _, err := os.Stat("."); err == nil {
		_ = x // want "identifier 'x' can be declared in the inner block to reduce its scope"
	}
	return nil
}

func UsedOnlyInsideFor() {
	v := 0
	for i := 0; i < 1; i++ {
		v = i // want "identifier 'v' can be declared in the inner block to reduce its scope"
		_ = v
	}
}

func MultiAssignSkipped() {
	a, err := os.ReadFile("f")
	_ = a
	if err != nil {
		_ = err
	}
}
