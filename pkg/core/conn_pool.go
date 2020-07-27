package core

import (
	"context"
	"errors"
	"sync"
)

type ConnectionFactory interface {
	Open(ctx context.Context, url string) (interface{}, error)
	Close(interface{}) error
}

//Due to multiple factors in play, it is possible that the application
//ends up connecting to a subset of the partitions. Hence this pool
//should create connections lazily
// todo: instead of taking locks on table. move to a volatile immutable map
type nonExpiryPool struct {
	table      map[string]interface{}
	mu, connMu sync.Mutex
	factory    ConnectionFactory
}

type ConnPool interface {
	Get(ctx context.Context, url string) (interface{}, error)
}

func NonExpiryPool(factory ConnectionFactory) *nonExpiryPool {
	return &nonExpiryPool{
		table:   make(map[string]interface{}),
		factory: factory,
	}
}

func (m *nonExpiryPool) Get(ctx context.Context, url string) (interface{}, error) {
	m.mu.Lock()
	conn, exists := m.table[url]
	m.mu.Unlock()

	if exists {
		return conn, nil
	}
	var existingConnection interface{}

	//to avoid multiple people from creating a connection
	//to the same url, take a lock on factory
	m.connMu.Lock()
	{
		var err error
		if conn, err = m.factory.Open(ctx, url); err != nil {
			m.connMu.Unlock()
			return nil, err
		}

		m.mu.Lock()
		{
			existingConnection, exists = m.table[url]
			if !exists {
				m.table[url] = conn
			}
		}
		m.mu.Unlock()
	}
	m.connMu.Unlock()

	// this close handling will happen after releasing the locks
	// on both table and factory
	if exists {
		if err := m.factory.Close(conn); err != nil {
			return nil, errors.New(
				"BadState: created an already existing connection and also failed to close it")
		}
		return existingConnection, nil
	}

	return conn, nil
}
