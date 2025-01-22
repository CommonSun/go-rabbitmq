package rabbitmq

import (
	"errors"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wagslane/go-rabbitmq/internal/channelmanager"
	"github.com/wagslane/go-rabbitmq/internal/connectionmanager"
)

type Manager struct {
	chanManager                *channelmanager.ChannelManager
	connManager                *connectionmanager.ConnectionManager
	reconnectErrCh             <-chan error
	closeConnectionToManagerCh chan<- struct{}

	disableDueToBlocked   bool
	disableDueToBlockedMu *sync.RWMutex

	options ManagerOptions
}

func NewManager(conn *Conn, optionFuncs ...func(*ManagerOptions)) (*Manager, error) {
	defaultOptions := getDefaultManagerOptions()
	options := &defaultOptions
	for _, optionFunc := range optionFuncs {
		optionFunc(options)
	}

	if conn.connectionManager == nil {
		return nil, errors.New("connection manager can't be nil")
	}

	chanManager, err := channelmanager.NewChannelManager(conn.connectionManager, options.Logger, conn.connectionManager.ReconnectInterval)
	if err != nil {
		return nil, err
	}

	reconnectErrCh, closeCh := chanManager.NotifyReconnect()

	manager := &Manager{
		chanManager:                chanManager,
		connManager:                conn.connectionManager,
		reconnectErrCh:             reconnectErrCh,
		closeConnectionToManagerCh: closeCh,
		disableDueToBlocked:        false,
		disableDueToBlockedMu:      &sync.RWMutex{},
		options:                    *options,
	}

	if err := manager.startup(); err != nil {
		return nil, err
	}

	go func() {
		for err := range manager.reconnectErrCh {
			manager.options.Logger.Infof("successful manager recovery from: %v", err)
			if err := manager.startup(); err != nil {
				manager.options.Logger.Fatalf("manager closing, unable to recover error on startup after cancel or close: %v", err)
				return
			}
		}
	}()

	return manager, nil
}

func (manager *Manager) Close() {
	err := manager.chanManager.Close()
	if err != nil {
		manager.options.Logger.Warnf("error closing manager channel: %v", err)
	}
	manager.options.Logger.Infof("closing manager...")

	go func() {
		manager.closeConnectionToManagerCh <- struct{}{}
	}()
}

func (manager *Manager) startup() error {
	for _, queueOptions := range manager.options.QueueOptions {
		if err := manager.DeclareQueue(queueOptions); err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}
	}

	for _, exchangeOptions := range manager.options.ExchangeOptions {
		if err := manager.DeclareExchange(exchangeOptions); err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	for _, consumerOptions := range manager.options.ConsumerOptions {
		if err := manager.DeclareBindings(consumerOptions); err != nil {
			return fmt.Errorf("failed to declare binding: %w", err)
		}
	}

	go manager.startNotifyBlockedHandler()
	return nil
}

func (manager *Manager) startNotifyBlockedHandler() {
	blockings := manager.connManager.NotifyBlockedSafe(make(chan amqp.Blocking))
	manager.disableDueToBlockedMu.Lock()
	manager.disableDueToBlocked = false
	manager.disableDueToBlockedMu.Unlock()

	for b := range blockings {
		manager.disableDueToBlockedMu.Lock()
		if b.Active {
			manager.options.Logger.Warnf("pausing manager due to TCP blocking from server")
			manager.disableDueToBlocked = true
		} else {
			manager.disableDueToBlocked = false
			manager.options.Logger.Warnf("resuming manager due to TCP blocking from server")
		}
		manager.disableDueToBlockedMu.Unlock()
	}
}
