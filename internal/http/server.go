package http

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"panic/internal/config"
	"panic/internal/db"
	"panic/internal/service"
	"panic/web"
)

const (
	actorKey      = "actor"
	csrfContext   = "csrf"
	sessionCookie = "panic_session"
	csrfCookie    = "panic_csrf"
)

type Server struct {
	cfg      config.Config
	db       *gorm.DB
	auth     *service.Auth
	catalog  *service.Catalog
	orders   *service.Orders
	payments *service.Payments
	tmpl     *template.Template
	files    http.Handler
	secure   bool
	engine   *gin.Engine
}

func New(cfg config.Config, conn *gorm.DB) (*Server, error) {
	tmpl, err := ParseTemplates()
	if err != nil {
		return nil, err
	}
	static, err := fs.Sub(web.Files, "static")
	if err != nil {
		return nil, err
	}
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	s := &Server{
		cfg:      cfg,
		db:       conn,
		auth:     service.NewAuth(conn, cfg.JWTSecret, cfg.JWTExpiresInHours),
		catalog:  service.NewCatalog(conn),
		orders:   service.NewOrders(conn),
		payments: service.NewPayments(conn, cfg.MidtransServerKey, cfg.MidtransClientKey, cfg.MidtransIsProduction),
		tmpl:     tmpl,
		files:    http.FileServer(http.FS(static)),
		secure:   cfg.SecureCookie(),
	}
	s.engine = s.routes()
	return s, nil
}

func Run(cfg config.Config, conn *gorm.DB) error {
	s, err := New(cfg, conn)
	if err != nil {
		return err
	}
	if err := s.engine.SetTrustedProxies(nil); err != nil {
		return err
	}
	addr := ":" + cfg.HTTPPort
	log.Printf("panic listening on %s", addr)
	return s.engine.Run(addr)
}

func (s *Server) routes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if s.cfg.AppEnv != "production" {
		r.Use(gin.Logger())
	}
	r.Use(s.securityHeaders(), s.limitBody(), s.authenticate(), s.csrfGuard())

	r.GET("/health", s.health)
	r.GET("/static/*filepath", s.static)

	r.GET("/login", s.loginForm)
	r.POST("/login", s.login)
	r.POST("/logout", s.logout)

	r.GET("/bayar/:invoice", s.payment)
	r.POST("/api/v1/midtrans/notification", s.notification)

	admin := r.Group("/admin", s.requireUser())
	admin.GET("", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/admin/products") })
	admin.GET("/products", s.productList)
	admin.GET("/products/new", s.productNew)
	admin.POST("/products", s.productCreate)
	admin.GET("/products/:id/edit", s.productEdit)
	admin.POST("/products/:id", s.productUpdate)
	admin.POST("/products/:id/delete", s.productDelete)
	admin.GET("/orders", s.orderList)

	r.GET("/", s.home)
	r.GET("/:seller", s.store)
	r.GET("/:seller/:product", s.landing)
	r.GET("/:seller/:product/checkout", s.checkoutForm)
	r.POST("/:seller/:product/checkout", s.checkoutSubmit)
	r.NoRoute(s.notFound)
	return r
}

func (s *Server) health(c *gin.Context) {
	if err := db.Ping(s.db); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) static(c *gin.Context) {
	if s.cfg.AppEnv != "production" {
		c.Header("Cache-Control", "no-cache")
	}
	req := c.Request.Clone(c.Request.Context())
	req.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/static")
	if req.URL.Path == "" {
		req.URL.Path = "/"
	}
	s.files.ServeHTTP(c.Writer, req)
}

func (s *Server) securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/admin") || path == "/login" || strings.HasPrefix(path, "/bayar") {
			c.Header("X-Robots-Tag", "noindex")
		}
		c.Next()
	}
}

func (s *Server) limitBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	}
}

func (s *Server) publicBase(c *gin.Context) string {
	if s.cfg.AppURL != "" {
		return s.cfg.AppURL
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func (s *Server) setCookie(c *gin.Context, name, value string, maxAge int, httpOnly bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", s.secure, httpOnly)
}

func (s *Server) currentActor(c *gin.Context) (service.Actor, bool) {
	value, ok := c.Get(actorKey)
	if !ok {
		return service.Actor{}, false
	}
	actor, ok := value.(service.Actor)
	return actor, ok
}

func (s *Server) authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookie)
		if err == nil && token != "" {
			if actor, err := s.auth.ActorFromToken(token); err == nil {
				c.Set(actorKey, actor)
			}
		}
		c.Next()
	}
}

func (s *Server) requireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := s.currentActor(c); ok {
			c.Next()
			return
		}
		c.Redirect(http.StatusSeeOther, "/login?next="+url.QueryEscape(safeNext(c.Request.URL.Path)))
		c.Abort()
	}
}
