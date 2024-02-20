// MIT License
//
// Copyright (c) 2020 The vine Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package web is a web dashboard
package web

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/cmd/vine/app/api/handler"
	"github.com/vine-io/vine/cmd/vine/client/resolver/web"
	"github.com/vine-io/vine/core/client/selector"
	"github.com/vine-io/vine/core/registry"
	"github.com/vine-io/vine/lib/api/server"
	"github.com/vine-io/vine/lib/api/server/cors"
	httpapi "github.com/vine-io/vine/lib/api/server/http"
	"github.com/vine-io/vine/lib/cmd"
	log "github.com/vine-io/vine/lib/logger"
	"github.com/vine-io/vine/util/helper"
	"github.com/vine-io/vine/util/namespace"
	"github.com/vine-io/vine/util/snaker"
	"github.com/vine-io/vine/util/stats"
	"golang.org/x/net/publicsuffix"
)

// Meta Fields of vine web
var (
	// Name default server name
	Name = "go.vine.web"
	// Address default address to bind to
	Address = ":8082"
	// Namespace the namespace to serve
	// Example:
	// Namespace + /[Service]/foo/bar
	// Host: Namespace.Service Endpoint: /foo/bar
	Namespace = "go.vine"
	Type      = "web"
	Resolver  = "path"
	// BasePathHeader base path sent to web service.
	// This is stripped from the request path
	// Allows the web service to define absolute paths
	BasePathHeader = "X-Vine-Web-Base-Path"
	statsURL       string
	loginURL       string

	// Host name the web dashboard is served on
	Host, _ = os.Hostname()
)

type service struct {
	app *gin.Engine
	// registry we use
	registry registry.Registry
	// the resolver
	resolver *web.Resolver
	// the namespace resolver
	nsResolver *namespace.Resolver
	// the proxy server
	prx *proxy
}

type reg struct {
	registry.Registry

	sync.RWMutex
	lastPull time.Time
	services []*registry.Service
}

// Handle serves the web dashboard and proxies where appropriate
func (s *service) Handle(c *gin.Context) {
	//host := string(c.Request().Host())
	//if len(c.Request().Host()) == 0 {
	//	r.URL.Host = r.Host
	//}
	//
	//if len(r.URL.Scheme) == 0 {
	//	r.URL.Scheme = "http"
	//}
	//
	//// no host means dashboard
	//host := r.URL.Hostname()
	//if len(host) == 0 {
	//	h, _, err := net.SplitHostPort(r.Host)
	//	if err != nil && strings.Contains(err.Error(), "missing port in address") {
	//		host = r.Host
	//	} else if err == nil {
	//		host = h
	//	}
	//}
	//
	//// check again
	//if len(host) == 0 {
	//	s.Router.ServeHTTP(w, r)
	//	return
	//}
	//
	//// check based on host set
	//if len(Host) > 0 && Host == host {
	//	s.Router.ServeHTTP(w, r)
	//	return
	//}
	//
	//// an ip instead of hostname means dashboard
	//ip := net.ParseIP(host)
	//if ip != nil {
	//	s.Router.ServeHTTP(w, r)
	//	return
	//}
	//
	//// namespace matching host means dashboard
	//parts := strings.Split(host, ".")
	//reverse(parts)
	//namespace := strings.Join(parts, ".")
	//
	//// replace mu since we know its ours
	//if strings.HasPrefix(namespace, "mu.vine") {
	//	namespace = strings.Replace(namespace, "mu.vine", "go.vine", 1)
	//}
	//
	//// web dashboard if namespace matches
	//if namespace == Namespace+"."+Type {
	//	s.Router.ServeHTTP(w, r)
	//	return
	//}
	//
	//// if a host has no subdomain serve dashboard
	//v, err := publicsuffix.EffectiveTLDPlusOne(host)
	//if err != nil || v == host {
	//	s.Router.ServeHTTP(w, r)
	//	return
	//}
	//
	//// check if its a web request
	//if _, _, isWeb := s.resolver.Info(r); isWeb {
	//	s.Router.ServeHTTP(w, r)
	//	return
	//}

	// otherwise serve the proxy
	s.prx.Handler(c)
}

// proxy is a http reverse proxy
func (s *service) proxy() *proxy {
	director := func(c *gin.Context) {
		//kill := func() {
		//	r.URL.Host = ""
		//	r.URL.Path = ""
		//	r.URL.Scheme = ""
		//	r.Host = ""
		//	r.RequestURI = ""
		//}
		//
		//// check to see if the endpoint was encoded in the request context
		//// by the auth wrapper
		//var endpoint *res.Endpoint
		//if val, ok := (r.Context().Value(res.Endpoint{})).(*res.Endpoint); ok {
		//	endpoint = val
		//}
		//
		//// TODO: better error handling
		//var err error
		//if endpoint == nil {
		//	if endpoint, err = s.resolver.Resolve(r); err != nil {
		//		log.Errorf("Failed to resolve url: %v: %v\n", r.URL, err)
		//		kill()
		//		return
		//	}
		//}
		//
		//r.Header.Set(BasePathHeader, "/"+endpoint.Name)
		//r.URL.Host = endpoint.Host
		//r.URL.Path = endpoint.Path
		//r.URL.Scheme = "http"
		//r.Host = r.URL.Host
	}

	return &proxy{
		//Router:   &httputil.ReverseProxy{Director: director},
		Director: director,
	}
}

func format(v *registry.Value) string {
	if v == nil || len(v.Values) == 0 {
		return "{}"
	}
	var f []string
	for _, k := range v.Values {
		f = append(f, formatEndpoint(k, 0))
	}
	return fmt.Sprintf("{\n%s}", strings.Join(f, ""))
}

func formatEndpoint(v *registry.Value, r int) string {
	// default format is tabbed plus the value plus new line
	fparts := []string{"", "%s %s", "\n"}
	for i := 0; i < r+1; i++ {
		fparts[0] += "\t"
	}
	// its just a primitive of sorts so return
	if len(v.Values) == 0 {
		return fmt.Sprintf(strings.Join(fparts, ""), snaker.CamelToSnake(v.Name), v.Type)
	}

	// this thing has more things, it's complex
	fparts[1] += " {"

	vals := []interface{}{snaker.CamelToSnake(v.Name), v.Type}

	for _, val := range v.Values {
		fparts = append(fparts, "%s")
		vals = append(vals, formatEndpoint(val, r+1))
	}

	// at the end
	l := len(fparts) - 1
	for i := 0; i < r+1; i++ {
		fparts[l] += "\t"
	}
	fparts = append(fparts, "}\n")

	return fmt.Sprintf(strings.Join(fparts, ""), vals...)
}

func faviconHandler(c *gin.Context) {
	return
}

func (s *service) indexHandler(c *gin.Context) {
	cors.SetHeaders(c.Writer, c.Request, &cors.Config{})

	if c.Request.Method == "OPTIONS" {
		return
	}

	services, err := s.registry.ListServices(c)
	if err != nil {
		log.Errorf("Error listing services: %v", err)
	}

	type webService struct {
		Name string
		Link string
		Icon string // TODO: lookup icon
	}

	// if the resolver is subdomain, we will need the domain
	domain, _ := publicsuffix.EffectiveTLDPlusOne(c.Request.URL.Hostname())

	var webServices []webService
	for _, svc := range services {
		// not a web app
		comps := strings.Split(svc.Name, ".web.")
		if len(comps) == 1 {
			continue
		}
		name := comps[1]

		link := fmt.Sprintf("/%v/", name)
		if Resolver == "subdomain" && len(domain) > 0 {
			link = fmt.Sprintf("https://%v.%v", name, domain)
		}

		// in the case of 3 letter things e.g m3o convert to M3O
		if len(name) <= 3 && strings.ContainsAny(name, "012345789") {
			name = strings.ToUpper(name)
		}

		webServices = append(webServices, webService{Name: name, Link: link})
	}

	sort.Slice(webServices, func(i, j int) bool { return webServices[i].Name < webServices[j].Name })

	type templateData struct {
		HasWebServices bool
		WebServices    []webService
	}

	data := templateData{len(webServices) > 0, webServices}
	s.render(c, indexTemplate, data)
}

func (s *service) registryHandler(c *gin.Context) {
	//vars := mux.Vars(c)
	//svc := vars["name"]
	//
	//if len(svc) > 0 {
	//	sv, err := s.registry.GetService(svc, registry.GetContext(r.Context()))
	//	if err != nil {
	//		http.Error(w, "Error occurred:"+err.Error(), 500)
	//		return
	//	}
	//
	//	if len(sv) == 0 {
	//		http.Error(w, "Not found", 404)
	//		return
	//	}
	//
	//	if r.Header.Get("Content-Type") == "application/json" {
	//		b, err := json.Marshal(map[string]interface{}{
	//			"services": s,
	//		})
	//		if err != nil {
	//			http.Error(w, "Error occurred:"+err.Error(), 500)
	//			return
	//		}
	//		w.Header().Set("Content-Type", "application/json")
	//		w.Write(b)
	//		return
	//	}
	//
	//	s.render(c, serviceTemplate, sv)
	//	return
	//}
	//
	//services, err := s.registry.ListServices(registry.ListContext(r.Context()))
	//if err != nil {
	//	log.Errorf("Error listing services: %v", err)
	//}
	//
	//sort.Sort(sortedServices{services})
	//
	//if r.Header.Get("Content-Type") == "application/json" {
	//	b, err := json.Marshal(map[string]interface{}{
	//		"services": services,
	//	})
	//	if err != nil {
	//		http.Error(w, "Error occurred:"+err.Error(), 500)
	//		return
	//	}
	//	w.Header().Set("Content-Type", "application/json")
	//	w.Write(b)
	//	return
	//}

	//return s.render(c, registryTemplate, services)
}

func (s *service) callHandler(c *gin.Context) {
	//services, err := s.registry.ListServices(registry.ListContext(c.Context()))
	//if err != nil {
	//	log.Errorf("Error listing services: %v", err)
	//}
	//
	//sort.Sort(sortedServices{services})
	//
	//serviceMap := make(map[string][]*registry.Endpoint)
	//for _, service := range services {
	//	if len(service.Endpoints) > 0 {
	//		serviceMap[service.Name] = service.Endpoints
	//		continue
	//	}
	//	// lookup the endpoints otherwise
	//	s, err := s.registry.GetService(service.Name, registry.GetContext(r.Context()))
	//	if err != nil {
	//		continue
	//	}
	//	if len(s) == 0 {
	//		continue
	//	}
	//	serviceMap[service.Name] = s[0].Endpoints
	//}
	//
	//if r.Header.Get("Content-Type") == "application/json" {
	//	b, err := json.Marshal(map[string]interface{}{
	//		"services": services,
	//	})
	//	if err != nil {
	//		http.Error(w, "Error occurred:"+err.Error(), 500)
	//		return
	//	}
	//	w.Header().Set("Content-Type", "application/json")
	//	w.Write(b)
	//	return
	//}
	//
	//return s.render(c, callTemplate, serviceMap)
}

func (s *service) render(c *gin.Context, tmpl string, data interface{}) {
	//t, err := template.New("template").Funcs(template.FuncMap{
	//	"format": format,
	//	"Title":  strings.Title,
	//	"First": func(s string) string {
	//		if len(s) == 0 {
	//			return s
	//		}
	//		return strings.Title(string(s[0]))
	//	},
	//}).Parse(layoutTemplate)
	//if err != nil {
	//	http.Error(w, "Error occurred:"+err.Error(), 500)
	//	return
	//}
	//t, err = t.Parse(tmpl)
	//if err != nil {
	//	http.Error(w, "Error occurred:"+err.Error(), 500)
	//	return
	//}
	//
	//// If the user is logged in, render Account instead of Login
	//loginTitle := "Login"
	//user := ""
	//
	//if c, err := r.Cookie(inauth.TokenCookieName); err == nil && c != nil {
	//	token := strings.TrimPrefix(c.Value, inauth.TokenCookieName+"=")
	//	if acc, err := s.auth.Inspect(token); err == nil {
	//		loginTitle = "Account"
	//		user = acc.ID
	//	}
	//}
	//
	//if err := t.ExecuteTemplate(w, "layout", map[string]interface{}{
	//	"LoginTitle": loginTitle,
	//	"LoginURL":   loginURL,
	//	"StatsURL":   statsURL,
	//	"Results":    data,
	//	"User":       user,
	//}); err != nil {
	//	http.Error(w, "Error occurred:"+err.Error(), 500)
	//}
}

func Run(c *cobra.Command, svcOpts ...vine.Option) error {

	ctx := c.PersistentFlags()
	if name, _ := ctx.GetString("server-name"); len(name) > 0 {
		Name = name
	}
	if addr, _ := ctx.GetString("address"); len(addr) > 0 {
		Address = addr
	}
	if r, _ := ctx.GetString("resolver"); len(r) > 0 {
		Resolver = r
	}
	if t, _ := ctx.GetString("type"); len(t) > 0 {
		Type = t
	}
	if ns, _ := ctx.GetString("namespace"); len(ns) > 0 {
		// remove the service type from the namespace to allow for
		// backwards compatability
		Namespace = strings.TrimSuffix(ns, "."+Type)
	}

	// service opts
	svcOpts = append(svcOpts, vine.Name(Name))

	// Initialize Server
	svc := vine.NewService(svcOpts...)

	reg := &reg{Registry: *cmd.DefaultOptions().Registry}

	s := &service{
		app:      gin.New(),
		registry: reg,
		// our internal resolver
		resolver: &web.Resolver{
			// Default to type path
			Type:      Resolver,
			Namespace: namespace.NewResolver(Type, Namespace).ResolveWithType,
			Selector: selector.NewSelector(
				selector.Registry(reg),
			),
		},
	}

	if b, _ := ctx.GetBool("enable-stats"); b {
		statsURL = "/stats"
		st := stats.New()
		s.app.Any("/stats", st.StatsHandler)
		st.Start()
		defer st.Stop()
	}

	// create the proxy
	p := s.proxy()

	// the web handler itself
	s.app.Any("/favicon.ico", faviconHandler)
	s.app.Any("/client", s.callHandler)
	s.app.Any("/services", s.registryHandler)
	s.app.Any("/service/{name}", s.registryHandler)
	s.app.Any("/rpc", handler.RPC)
	s.app.Any("/{service:[a-zA-Z0-9]+}", p.Handler)
	s.app.Any("/", s.indexHandler)

	// insert the proxy
	s.prx = p

	var opts []server.Option

	if b, _ := ctx.GetBool("enable-tls"); b {
		config, err := helper.TLSConfig(c)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		opts = append(opts, server.EnableTLS(true))
		opts = append(opts, server.TLSConfig(config))
	}

	// create the namespace resolver and the auth wrapper
	s.nsResolver = namespace.NewResolver(Type, Namespace)

	// create the service and add the auth wrapper
	server := httpapi.NewServer(Address)

	server.Init(opts...)
	server.Handle("/", s.app)

	if err := server.Start(); err != nil {
		return err
	}

	// Run server
	if err := svc.Run(); err != nil {
		return err
	}

	if err := server.Stop(); err != nil {
		return err
	}

	return nil
}

// Commands for `vine web`
func Commands(options ...vine.Option) []*cobra.Command {
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Run the web dashboard",
		RunE: func(c *cobra.Command, args []string) error {
			return Run(c, options...)
		},
	}
	webCmd.PersistentFlags().String("address", "", "Set the web UI address e.g 0.0.0.0:8082")
	webCmd.PersistentFlags().String("namespace", "", "Set the namespace used by the Web proxy e.g. com.example.web")
	webCmd.PersistentFlags().String("resolver", "", "Set the resolver to route to services e.g path, domain")
	webCmd.PersistentFlags().String("auth-login-url", "", "The relative URL where a user can login")

	return []*cobra.Command{webCmd}
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
