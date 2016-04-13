// code base on https://github.com/martini-contrib/web
// just for myself.
// Insion Ng
// Sun Apr 10 11:25 AM
package selfor

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// Selfor 引用的Macaron必须与其他中间件中引用的版本保持一致
	// 不然会无法匹配到*macaron.Context，导致程序崩溃
	"gopkg.in/macaron.v1" //当前使用的版本，如果需要编写中间件请使用这个版本
	//"github.com/Unknwon/macaron"//请勿在其他中间件中使用此版本
	//"github.com/go-macaron/macaron"//请勿在其他中间件中使用此版本
)

// A Context object is created for every incoming HTTP request, and is
// passed to handlers as an optional first argument. It provides information
// about the request, including the http.Request object, the GET and POST params,
// and acts as a Writer for the response.
type Context struct {
	*macaron.Context
	Request      *http.Request
	Params       map[string]string
	CookieSecret string
	Response     http.ResponseWriter
}

// if cookie secret is set to nil, then SetSecureCookie use default Secure.
func Selfor(secret []byte) macaron.Handler {
	return func(c *macaron.Context, w http.ResponseWriter, req *http.Request) {
		_secret := string(secret)
		if len(_secret) == 0 {
			h := hmac.New(md5.New, []byte("Selfor Marcaron!"))
			h.Write([]byte(c.Req.UserAgent()))
			_secret = hex.EncodeToString(h.Sum(nil))
		}

		//ctx := &Context{ req, map[string]string{}, _secret, w}

		ctx := &Context{
			//c, //Caution, cannot put the c here!
			Request:      req,
			Params:       map[string]string{},
			CookieSecret: _secret,
			Response:     w,
		}

		//set some default headers
		tm := time.Now().UTC()

		//ignore errors from ParseForm because it's usually harmless.
		req.ParseForm()
		if len(req.Form) > 0 {
			for k, v := range req.Form {
				ctx.Params[k] = v[0]
			}
		}
		ctx.SetHeader("Date", webTime(tm), true)
		//Set the default content-type
		ctx.SetHeader("Content-Type", "text/html; charset=utf-8", true)

		// set macaron context for selfor.Context
		c.Map(ctx)

	}

}

func Classic(o io.Writer) *macaron.Macaron {

	if o == nil {
		o = os.Stdout
	}

	var r = macaron.NewRouter()
	var m = macaron.NewWithLogger(o)
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public"))
	m.Map(r)
	m.Action(r.Handle)
	m.Use(Selfor(nil))
	return m

}

// internal utility methods
func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

// WriteString writes string data into the response object.
func (ctx *Context) WriteString(content string) {
	ctx.Response.Write([]byte(content))
}

// Abort is a helper method that sends an HTTP header and an optional
// body. It is useful for returning 4xx or 5xx errors.
// Once it has been called, any return value from the handler will
// not be written to the response.
func (ctx *Context) Abort(status int, body string) {
	ctx.Response.WriteHeader(status)
	ctx.Response.Write([]byte(body))
}

// Redirect is a helper method for 3xx redirects.
func (ctx *Context) Redirect(status int, url_ string) {
	ctx.Response.Header().Set("Location", url_)
	ctx.Response.WriteHeader(status)
	ctx.Response.Write([]byte("Redirecting to: " + url_))
}

// Notmodified writes a 304 HTTP response
func (ctx *Context) NotModified() {
	ctx.Response.WriteHeader(304)
}

// NotFound writes a 404 HTTP response
func (ctx *Context) NotFound(message string) {
	ctx.Response.WriteHeader(404)
	ctx.Response.Write([]byte(message))
}

//Unauthorized writes a 401 HTTP response
func (ctx *Context) Unauthorized() {
	ctx.Response.WriteHeader(401)
}

//Forbidden writes a 403 HTTP response
func (ctx *Context) Forbidden() {
	ctx.Response.WriteHeader(403)
}

// ContentType sets the Content-Type header for an HTTP response.
// For example, ctx.ContentType("json") sets the content-type to "application/json"
// If the supplied value contains a slash (/) it is set as the Content-Type
// verbatim. The return value is the content type as it was
// set, or an empty string if none was found.
func (ctx *Context) ContentType(val string) string {
	var ctype string
	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	if ctype != "" {
		ctx.Response.Header().Set("Content-Type", ctype)
	}
	return ctype
}

// SetHeader sets a response header. If `unique` is true, the current value
// of that header will be overwritten . If false, it will be appended.
func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.Response.Header().Set(hdr, val)
	} else {
		ctx.Response.Header().Add(hdr, val)
	}
}

// SetCookie adds a cookie header to the response.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}

func getCookieSig(key string, val []byte, timestamp string) string {
	hm := hmac.New(sha1.New, []byte(key))

	hm.Write(val)
	hm.Write([]byte(timestamp))

	hex := fmt.Sprintf("%02x", hm.Sum(nil))
	return hex
}

// NewCookie is a helper method that returns a new http.Cookie object.
// Duration is specified in seconds. If the duration is zero, the cookie is permanent.
// This can be used in conjunction with ctx.SetCookie.
func NewCookie(name string, value string, age int64) *http.Cookie {
	var utctime time.Time
	if age == 0 {
		// 2^31 - 1 seconds (roughly 2038)
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	return &http.Cookie{Name: name, Value: value, Expires: utctime}
}

func (ctx *Context) SetSecureCookie(name string, val string, age int64) {
	//base64 encode the val
	if len(ctx.CookieSecret) == 0 {
		return
	}
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(val))
	encoder.Close()
	vs := buf.String()
	vb := buf.Bytes()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := getCookieSig(ctx.CookieSecret, vb, timestamp)
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	ctx.SetCookie(NewCookie(name, cookie, age))
}

func (ctx *Context) GetSecureCookie(name string) (string, bool) {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name != name {
			continue
		}

		parts := strings.SplitN(cookie.Value, "|", 3)

		val := parts[0]
		timestamp := parts[1]
		sig := parts[2]

		if getCookieSig(ctx.CookieSecret, []byte(val), timestamp) != sig {
			return "", false
		}

		ts, _ := strconv.ParseInt(timestamp, 0, 64)

		if time.Now().Unix()-31*86400 > ts {
			return "", false
		}

		buf := bytes.NewBufferString(val)
		encoder := base64.NewDecoder(base64.StdEncoding, buf)

		res, _ := ioutil.ReadAll(encoder)
		return string(res), true
	}
	return "", false
}
