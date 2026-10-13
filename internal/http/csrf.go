package http

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) csrfGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(csrfCookie)
		if err != nil || len(token) != 64 {
			token, err = newCSRFToken()
			if err != nil {
				c.String(http.StatusInternalServerError, "Gagal menyiapkan formulir")
				c.Abort()
				return
			}
			s.setCookie(c, csrfCookie, token, 7*24*3600, true)
		}
		c.Set(csrfContext, token)
		if c.Request.Method == http.MethodPost && c.Request.URL.Path != "/api/v1/midtrans/notification" {
			_ = c.Request.ParseForm()
			if !sameToken(c.PostForm("csrf"), token) {
				s.renderStatus(c, http.StatusForbidden, "page-message", messageView{
					view:    s.baseView(c, "Formulir ditolak"),
					Message: "Sesi formulir habis. Kembali dan muat ulang halaman.",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func (s *Server) csrfToken(c *gin.Context) string {
	value, _ := c.Get(csrfContext)
	token, _ := value.(string)
	return token
}

func newCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func sameToken(got, want string) bool {
	if len(got) != len(want) || len(got) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
