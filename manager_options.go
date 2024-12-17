package rabbitmq

type ManagerOptions struct {
	ExchangeOptions []ExchangeOptions
	QueueOptions    []QueueOptions
	ConsumerOptions []ConsumerOptions
	Logger          Logger
}

func getDefaultManagerOptions() ManagerOptions {
	return ManagerOptions{
		Logger:          stdDebugLogger{},
		ExchangeOptions: make([]ExchangeOptions, 0),
		QueueOptions:    make([]QueueOptions, 0),
		ConsumerOptions: make([]ConsumerOptions, 0),
	}
}

// WithTopologyOptionsLogging sets logging to true on the topology options
func WithManagerOptionsLogging(options *ManagerOptions) {
	options.Logger = &stdDebugLogger{}
}

// WithManagerOptionsLogger sets logging to a custom interface.
// Use WithManagerOptionsLogging to just log to stdout.
func WithManagerOptionsLogger(log Logger) func(options *ManagerOptions) {
	return func(options *ManagerOptions) {
		options.Logger = log
	}
}

func WithManagerOptionsExchange(exchange ExchangeOptions) func(*ManagerOptions) {
	return func(options *ManagerOptions) {
		options.ExchangeOptions = append(options.ExchangeOptions, exchange)
	}
}

func ManagerDefaultExchangeOptions() ExchangeOptions {
	return getDefaultExchangeOptions()
}

func WithManagerOptionsQueue(queue QueueOptions) func(*ManagerOptions) {
	return func(options *ManagerOptions) {
		options.QueueOptions = append(options.QueueOptions, queue)
	}
}

func WithManagerOptionsConsumer(consumer ConsumerOptions) func(*ManagerOptions) {
	return func(options *ManagerOptions) {
		options.ConsumerOptions = append(options.ConsumerOptions, consumer)
	}
}
