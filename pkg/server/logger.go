package server

import (
	"net"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
)

func RequestLogger(log *logrus.Entry) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			resp := c.Response()
			start := time.Now()

			ra := req.RemoteAddr
			ra, port, _ := net.SplitHostPort(ra)
			if ip := req.Header.Get(echo.HeaderXRealIP); ip != "" {
				ra = ip
				port = ""
			} else if ip = req.Header.Get(echo.HeaderXForwardedFor); ip != "" {
				ra = ip
				port = ""
			}

			log = log.
				WithField("remote_address", ra).
				WithField("remote_port", port).
				WithField("url", req.URL).
				WithField("method", req.Method)

			err := next(c)
			end := time.Now()
			log = log.
				WithField("status", resp.Status).
				WithField("took", end.Sub(start))
			if rid := resp.Header().Get(echo.HeaderXRequestID); rid != "" {
				log = log.WithField("audit_id", rid)
			}
			if err != nil {
				c.Error(err)
				log.WithError(err).
					WithField("status", resp.Status).
					Error(http.StatusText(resp.Status))
				return err
			}
			log.Info(http.StatusText(resp.Status))
			return nil
		}
	}
}
