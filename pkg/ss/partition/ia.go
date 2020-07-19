package partition

type initActor struct {
	*actorImpl
}

func (ia *initActor) become() error {
	if ia.leader == nil {
		primea := primaryActor{ia.actorImpl}
		return primea.become()
	} else {
		fa := catchingUpActor{ia.actorImpl}
		return fa.become()
	}
}
