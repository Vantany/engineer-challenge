package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func New(limit int, window time.Duration) *Limiter {
	rl := &Limiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.window {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *Limiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}

	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	v.lastSeen = time.Now()
	v.count++
	return v.count <= rl.limit
}

type RouteLimiter struct {
	global *Limiter
	routes map[string]*Limiter
	mu     sync.RWMutex
}

func NewRouteLimiter(globalLimit int, globalWindow time.Duration) *RouteLimiter {
	return &RouteLimiter{
		global: New(globalLimit, globalWindow),
		routes: make(map[string]*Limiter),
	}
}

func (rl *RouteLimiter) AddRoute(route string, limit int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.routes[route] = New(limit, window)
}

func (rl *RouteLimiter) Allow(ip, route string) bool {
	rl.mu.RLock()
	routeLimiter, ok := rl.routes[route]
	rl.mu.RUnlock()

	if ok {
		if !routeLimiter.Allow(ip) {
			return false
		}
	}

	return rl.global.Allow(ip)
}
