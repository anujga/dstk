package partition

type initActor struct {
	*PartRange
}

func (ia *initActor) become() error {
	if ia.leader == nil {
		primea := primaryActor{ia.PartRange}
		return primea.become()
	} else {
		fa := catchingUpActor{ia.PartRange}
		return fa.become()
	}
}
