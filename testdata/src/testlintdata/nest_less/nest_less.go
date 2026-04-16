package nest_less

func bad() {
	cond1 := true
	cond2 := true
	cond3 := true
	if cond1 {
		if cond2 {
			if cond3 {
				if true { // want "reduce nesting: depth 4 exceeds allowed 3; consider returning early"
					_ = 1
				}
			}
		}
	}
}

func good() {
	cond1 := true
	cond2 := true
	cond3 := true
	if !cond1 {
		return
	}
	if !cond2 {
		return
	}
	if !cond3 {
		return
	}
	_ = 1
}
