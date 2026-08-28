package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	count     int
	windowEnd time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	limit   int
	window  time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		entries: make(map[string]*ipEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		rl.mu.Lock()
		now := time.Now()
		for ip, e := range rl.entries {
			if now.After(e.windowEnd) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, ok := rl.entries[ip]
	if !ok || now.After(e.windowEnd) {
		rl.entries[ip] = &ipEntry{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}
	if e.count >= rl.limit {
		return false
	}
	e.count++
	return true
}

var loginLimiter = newRateLimiter(5, 5*time.Minute)

// publicWriteLimiter protege las escrituras que no requieren sesión
// (registro, cotización, eventos). Es más holgado que el de login porque
// aquí el objetivo no es frenar fuerza bruta sino evitar que un script
// inunde la cola de moderación o la cuota de correo.
var publicWriteLimiter = newRateLimiter(20, time.Hour)

// RateLimit applies a per-IP rate limit to login endpoints (5 req / 5 min).
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !loginLimiter.allow(ip) {
			c.Header("Retry-After", "300")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "demasiados intentos, intenta de nuevo en 5 minutos",
			})
			return
		}
		c.Next()
	}
}

// eventLimiter es aparte y mucho más holgado: /api/events se dispara en cada
// vista de página, así que el límite estrecho de PublicWriteLimit devolvería
// 429 a usuarios normales y rompería la analítica. Aquí solo interesa cortar
// un flood evidente.
var eventLimiter = newRateLimiter(300, time.Hour)

// EventLimit limita el registro de eventos de analítica por IP.
func EventLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !eventLimiter.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "demasiados eventos"})
			return
		}
		c.Next()
	}
}

// PublicWriteLimit limita las escrituras públicas a 20 por hora y por IP.
// Holgado a propósito: varias empresas pueden compartir IP tras un NAT, y
// frente a un flood real la diferencia entre 10 y 20 es irrelevante.
func PublicWriteLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !publicWriteLimiter.allow(c.ClientIP()) {
			c.Header("Retry-After", "3600")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Recibimos demasiadas solicitudes desde esta conexión. Intenta de nuevo más tarde o escríbenos a contacto@puntofusion.cl.",
			})
			return
		}
		c.Next()
	}
}
