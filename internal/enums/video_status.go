package enums

type VideoStatus int

const (
	Pending VideoStatus = 1

	Processing VideoStatus = 2

	Completed VideoStatus = 3

	Failed VideoStatus = 4
)
