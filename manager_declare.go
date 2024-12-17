package rabbitmq

func (m *Manager) DeclareQueue(options QueueOptions) error {
	return declareQueue(m.chanManager, options)
}

func (m *Manager) DeclareExchange(options ExchangeOptions) error {
	return declareExchange(m.chanManager, options)
}

func (m *Manager) DeclareBindings(options ConsumerOptions) error {
	return declareBindings(m.chanManager, options)
}
