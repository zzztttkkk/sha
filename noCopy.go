package sha

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) UnLock() {}
