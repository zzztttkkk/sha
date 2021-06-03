package main

type Engine interface {
	Name() string
	HelloWorld(address string)
}
