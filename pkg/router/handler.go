package router

type HandlerFunc[T any] func(*T)
