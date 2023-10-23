package controllers

import "sort"

type controllersSort []string

func (x controllersSort) Len() int {
	return len(x)
}

func (x controllersSort) Less(i, j int) bool {
	a, b := x[i], x[j]
	if x[i][0] == '-' {
		a = x[i][1:]
	}
	if x[j][0] == '-' {
		b = x[j][1:]
	}
	return a < b
}

func (x controllersSort) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}
func (x controllersSort) Sort() {
	sort.Sort(x)
}
