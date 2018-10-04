// Copyright 2018 Jacob Dufault
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
)

const dismissCaptivePortalFilePath = "dismiss-captive-portal"
const dismissCaptivePortalEndPoint = "dismiss-captive-portal"

// HTML that is served when a captive portal is displayed.
var captivePortalHTML = fmt.Sprintf(`<html>
<head>
  <meta http-equiv="content-type" content="text/html;charset=utf-8">
  <title>Title</title>
</head>

<body>
  <h1>Captive portal dialog</h1>
  <p>This is a simple captive portal. To hide the captive portal, add a file
   	 called <code>allow</code> in the directory the proxy server is running.</p>
  <a href="/%s">Dismiss the captive portal</a>
</body>
</html>`, dismissCaptivePortalEndPoint)

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func readTrimmedFileContents(path string) string {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(fileContent))
}

func shouldShowPortal(req *http.Request) bool {
	if *captivePortal && !fileExists(dismissCaptivePortalFilePath) {
		if strings.Contains(req.URL.Path, dismissCaptivePortalEndPoint) {
			ioutil.WriteFile(dismissCaptivePortalFilePath, nil, 0664)
		}
		return true
	}

	return false
}

func portalUnauthorized(req *http.Request) *http.Response {
	return goproxy.NewResponse(req, goproxy.ContentTypeHtml, http.StatusFound, captivePortalHTML)
}

func portal() goproxy.ReqHandler {
	return goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if shouldShowPortal(req) {
			return nil, portalUnauthorized(req)
		}
		return req, nil
	})
}

func portalConnect() goproxy.HttpsHandler {
	return goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		if shouldShowPortal(ctx.Req) {
			ctx.Resp = portalUnauthorized(ctx.Req)
			return goproxy.RejectConnect, host
		}
		return goproxy.OkConnect, host
	})
}

var (
	basicAuth = flag.String("basic-auth", "",
		`Should the proxy require basic authentication? Format is user:password`)
	captivePortal = flag.Bool("captive-portal", false,
		fmt.Sprintf(`Should a captive portal be shown? The portal will be dismissed when a file
called "%s" is present in the working directory.`, dismissCaptivePortalFilePath))
	port = flag.Int("port", 8080,
		`What port should the proxy run on?`)
	verbose = flag.Bool("verbose", false,
		`Emit verbose output`)
)

func main() {
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	os.Remove(dismissCaptivePortalFilePath)

	if len(*basicAuth) > 0 {
		auth.ProxyBasic(proxy, "realm", func(user, password string) bool {
			actual := user + ":" + password
			if *verbose {
				fmt.Printf("BasicAuth: actual='%s', expected='%s'\n", actual, *basicAuth)
			}
			return actual == *basicAuth
		})
	}

	proxy.OnRequest().Do(portal())
	proxy.OnRequest().HandleConnect(portalConnect())

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	fmt.Printf("Serving on %s\n", addr)
	log.Fatalln(http.ListenAndServe(addr, proxy))
}
