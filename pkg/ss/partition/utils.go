package partition

func getActorFunction(msg BecomeMsg, ai *actorImpl) (fun func() error, err error) {
	ab := actorBase{
		id:             ai.Id(),
		logger:         ai.logger,
		smState:        ai.smState,
		mailBox:        ai.mailBox,
		consumer:       ai.consumer,
		partitionRpc:   ai.partitionRpc,
		currentDbState: StateFromString(ai.partition.GetCurrentState()),
	}
	switch msg.Target() {
	case Init:
		ia := initActor{ab}
		fun = ia.become
	case Primary:
		pa := primaryActor{ab}
		fun = pa.become
	case Proxy:
		bp := msg.(*BecomeProxy)
		pa := proxyActor{
			actorBase: ab,
			proxyTo:   bp.ProxyTo,
		}
		fun = pa.become
	case CatchingUp:
		bf := msg.(*BecomeCatchingUpActor)
		ca := catchingUpActor{
			actorBase:     ab,
			leaderMailbox: bf.LeaderMailbox,
		}
		fun = ca.become
	case Follower:
		fa := followingActor{ab}
		fun = fa.become
	}
	return
}
