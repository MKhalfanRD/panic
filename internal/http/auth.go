package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"panic/internal/service"
)

func (s *Server) loginForm(c *gin.Context) {
	if _, ok := s.currentActor(c); ok {
		c.Redirect(http.StatusSeeOther, "/admin/products")
		return
	}
	page := loginView{
		view:  s.baseView(c, "Masuk"),
		Next:  safeNext(c.Query("next")),
		Error: "",
	}
	page.Robots = "noindex, nofollow"
	page.DevHint = s.cfg.AppEnv == "development"
	s.render(c, "page-login", page)
}

func (s *Server) login(c *gin.Context) {
	next := safeNext(c.PostForm("next"))
	token, _, err := s.auth.Login(c.PostForm("email"), c.PostForm("password"))
	if err != nil {
		page := loginView{
			view:  s.baseView(c, "Masuk"),
			Next:  next,
			Error: "Email atau kata sandi salah.",
		}
		page.Robots = "noindex, nofollow"
		page.DevHint = s.cfg.AppEnv == "development"
		s.renderStatus(c, http.StatusUnprocessableEntity, "page-login", page)
		return
	}
	s.setCookie(c, sessionCookie, token, s.cfg.JWTExpiresInHours*3600, true)
	c.Redirect(http.StatusSeeOther, next)
}

func (s *Server) logout(c *gin.Context) {
	s.setCookie(c, sessionCookie, "", -1, true)
	c.Redirect(http.StatusSeeOther, "/login")
}

func (s *Server) mustActor(c *gin.Context) (service.Actor, bool) {
	actor, ok := s.currentActor(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/login")
		return service.Actor{}, false
	}
	return actor, true
}
