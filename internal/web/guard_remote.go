package web

func (g *guard) allowed(host string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hosts[host] || g.remote[host]
}

func (g *guard) isRemote(host string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.remote[host]
}

func (g *guard) setRemote(host string, enabled bool) {
	g.mu.Lock()
	if enabled {
		g.remote[host] = true
	} else {
		delete(g.remote, host)
	}
	g.mu.Unlock()
	if !enabled {
		g.codes.mu.Lock()
		clear(g.codes.codes)
		g.codes.mu.Unlock()
	}
}
