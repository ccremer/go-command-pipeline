package pipeline

func True() Predicate {
	return func(step Step) bool {
		return true
	}
}

func False() Predicate {
	return func(step Step) bool {
		return false
	}
}

func Bool(v bool) Predicate {
	return func(step Step) bool {
		return v
	}
}

func Not(predicate Predicate) Predicate {
	return func(step Step) bool {
		return !predicate(step)
	}
}

func And(p1, p2 Predicate) Predicate {
	return func(step Step) bool {
		return p1(step) && p2(step)
	}
}
