package suna

type noCopy struct{} //nolint:unused

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
