package model

type Priority string

const (
	P1 Priority = "P1"
	P2 Priority = "P2"
	P3 Priority = "P3"
)

func (p Priority) Label() string {
	switch p {
	case P1:
		return "P1"
	case P2:
		return "P2"
	default:
		return "P3"
	}
}

func (p Priority) Order() int {
	switch p {
	case P1:
		return 0
	case P2:
		return 1
	default:
		return 2
	}
}
