package eventmanager

type Handler[Message any] func(Message) (domainErr, infrastructureErr error)
