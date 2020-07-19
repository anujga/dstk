package partition


const (
	Init State = iota
	CatchingUp
	Follower
	Primary
	Proxy
	Completed
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
	case Completed:
		return "completed"
	default:
		return "invalid state"
	}
}
