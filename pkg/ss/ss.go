package ss

type MsgTrait interface {
	ReadOnly() bool
}


type Consumer interface {
	Process(msg MsgTrait) bool
}
