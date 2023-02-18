package enums

type ServerStatus int

const (
	Stopped ServerStatus = iota
	Preparing
	Running
)

func (s ServerStatus) String() string {
	switch s {
	case Stopped:
		return "Stopped"
	case Preparing:
		return "Preparing"
	case Running:
		return "Running"
	}
	return "unknown"
}
