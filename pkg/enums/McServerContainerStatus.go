package enums

type ServerStatus int

const (
	Prepared ServerStatus = iota
	Stopped
	Preparing
	Running
)

func (s ServerStatus) String() string {
	switch s {
	case Stopped:
		return "Stopped"
	case Preparing:
		return "Preparing"
	case Prepared:
		return "Prepared"
	case Running:
		return "Running"
	}
	return "unknown"
}
