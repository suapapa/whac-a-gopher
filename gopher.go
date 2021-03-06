package main

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
)

// EyeShape is position of eye of a gopher
type EyeShape int

// GopherStatus is status of gopher
type GopherStatus int

const (
	// EyeX means dead eye
	EyeX EyeShape = iota
	// EyeLeft means look left
	EyeLeft
	// EyeRight means look right
	EyeRight

	// Hide means gopher is in the hole
	Hide GopherStatus = iota
	// Peak means gopher peaks
	Peak
	// Dizzy means gopher is dizzed by hammer
	Dizzy
)

var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Gopher reperesent a gopher in a hole
type Gopher struct {
	status                GopherStatus
	eye                   EyeShape
	HeadCh, ButtCh        chan struct{}
	dizzyUntil, peakUntil time.Time
	rollEyeUntil          time.Time
	wg                    sync.WaitGroup
	sync.RWMutex          // Lock for status and eye
}

// NewGopher return adress of a Gopher
func NewGopher() *Gopher {
	return &Gopher{
		eye:    EyeX,
		HeadCh: make(chan struct{}, 1),
		ButtCh: make(chan struct{}, 1),
		status: Hide,
	}
}

// Start makes a gopher run
func (g *Gopher) Start(ctx context.Context) {
	log.Printf("start gopher: %p", g)
	g.wg.Add(2)
	go g.handleEvent(ctx)
	go g.updateStatus(ctx)
}

// Wait waits all goroutines are closed
func (g *Gopher) Wait() {
	g.wg.Wait()
}

func (g *Gopher) handleEvent(ctx context.Context) {
	defer g.wg.Done()
loop:
	select {
	case <-ctx.Done():
		return
	case <-g.HeadCh:
		if g.Status() != Hide {
			log.Println("ouch!! :%p", g)
			g.Lock()
			g.status = Dizzy
			g.eye = EyeX
			g.Unlock()
			g.dizzyUntil = time.Now().Add(500 * time.Millisecond)
		}
	case <-g.ButtCh:
		now := time.Now()
		if g.Status() == Hide {
			g.Lock()
			g.status = Peak
			g.eye = EyeShape(1 + r.Intn(2))
			g.rollEyeUntil = now.Add(time.Duration(r.Intn(500)) * time.Millisecond)
			g.peakUntil = now.Add(time.Duration(r.Intn(2000))*time.Millisecond + 100*time.Millisecond)
			g.Unlock()
		}
	default:
	}
	// runtime.Gosched()
	time.Sleep(20 * time.Millisecond)
	goto loop
}

func (g *Gopher) updateStatus(ctx context.Context) {
	defer g.wg.Done()
loop:
	select {
	case <-ctx.Done():
		return
	default:
	}

	switch g.Status() {
	case Dizzy:
		if time.Now().After(g.dizzyUntil) {
			g.Lock()
			g.status = Hide
			g.Unlock()
		}
	case Peak:
		now := time.Now()
		if now.After(g.peakUntil) {
			g.Lock()
			g.status = Hide
			g.Unlock()
		} else {
			if now.After(g.rollEyeUntil) {
				g.Lock()
				g.eye = EyeLeft + EyeShape(r.Intn(2))
				g.rollEyeUntil = now.Add(time.Duration(r.Intn(500)) * time.Millisecond)
				g.Unlock()
			}
		}
	default:
	}
	time.Sleep(20 * time.Millisecond)
	goto loop
}

// Eye returns currnet shape of eye of gopher
func (g *Gopher) Eye() EyeShape {
	g.RLock()
	defer g.RUnlock()
	return g.eye
}

// Status returns currnet status of gopher
func (g *Gopher) Status() GopherStatus {
	g.RLock()
	defer g.RUnlock()
	return g.status
}
