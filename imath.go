package main

func Abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

func Max(v1 int, v2 int) int {
	if v1 > v2 {
		return v1
	}

	return v2
}
