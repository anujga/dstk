package partition

type State int

const (
	Init State = iota
	CatchingUp
	Follower
	Primary
	Proxy
	Retired
	Invalid
)

func (s State) String() string {
	switch s {
	case Follower:
		return "follower"
	case CatchingUp:
		return "catchingup"
	case Primary:
		return "primary"
	case Proxy:
		return "proxy"
	//case Completed:
	//	return "completed"
	case Retired:
		return "retired"
	case Init:
		return "init"
	default:
		return "invalid state"
	}
}

func StateFromString(s string) State {
	switch s {
	case "proxy":
		return Proxy
	case "primary":
		return Primary
	case "follower":
		return Follower
	case "catchingup":
		return CatchingUp
	case "retired":
		return Retired
	default:
		return Invalid
	}
}
