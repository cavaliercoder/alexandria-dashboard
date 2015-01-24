/*
 * Alexandria CMDB - Open source config management database
 * Copyright (C) 2014  Ryan Armstrong <ryan@cavaliercoder.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var ApiAuthError = errors.New("API Authentication error")

type Controller struct {
	*revel.Controller
	authContext *AuthContext
}

type ApiOptions struct {
	Impersonate bool
	Body        io.Reader
	Selector    interface{}
	Limit       int
	Offset      int
}

// Check panics if the specified error is not nil.
func (c *Controller) Check(err error) {
	if err != nil {
		revel.ERROR.Panic(err)
	}
}

// ApiRequest executes a RESTful operation against the Alexandria API.
// The request can impersonate the authenticated user or be executed as the
// Dashboard API account.
// If a body is specified it is included as the request body.
func (c *Controller) ApiRequest(method string, path string, options ApiOptions) (*http.Response, error) {
	// Get API URL from configuration
	baseUrl, ok := revel.Config.String("api.url")
	if !ok {
		panic("API URL is not set")
	}

	// Strip API version prefix from the requested path as it should already be
	// present in the configured API URL
	r := regexp.MustCompile("^/api/v1")
	path = r.ReplaceAllString(path, "")

	// Build URL
	url := fmt.Sprintf("%s%s", baseUrl, path)
	method = strings.ToUpper(method)

	// Add authentication header
	var apiKey string
	if options.Impersonate {
		apiKey = c.Session["token"]
	} else {
		apiKey, ok = revel.Config.String("api.key")
		if !ok {
			revel.ERROR.Panic("API authentication key is not set")
		}
	}

	// Create a HTTP client that does not follow redirects
	// This allows 'Location' headers to be read
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("Never follow redirects")
		},
	}

	// Create the request
	req, err := http.NewRequest(method, url, options.Body)
	if err != nil {
		return nil, err
	}

	// Add request headers
	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Add("Content-type", "application/json")
	}

	if apiKey != "" {
		req.Header.Add("X-Auth-Token", apiKey)
	}

	// TODO: Update Dashboard user agent string with version info
	req.Header.Add("User-Agent", "Alexandria CMDB Dashboard")

	// Submit the request
	revel.TRACE.Printf("API started: %s %s", method, url)
	start := time.Now()
	res, err := client.Do(req)
	if res == nil {
		revel.ERROR.Panic("An error occurred communicating with backend services")
	}
	revel.TRACE.Printf("API finished: %s in %s", res.Status, time.Since(start))

	// Validate response
	if res.StatusCode == http.StatusUnauthorized {
		return res, ApiAuthError
	}

	return res, err
}

// ApiGet executes a RESTful GET operation on the Alexandria API and returns
// a pointer to the http.Response struct.
func (c *Controller) ApiGet(path string, options ApiOptions) (*http.Response, error) {
	return c.ApiRequest("GET", path, options)
}

// ApiGetString executes a RESTful GET operation on the Alexandria API and
// returns the response body as a string or an error.
func (c *Controller) ApiGetString(path string, options ApiOptions) (string, int, error) {
	res, err := c.ApiRequest("GET", path, options)
	if err != nil {
		return "", 0, err
	}

	// Read the response body into a string
	var bytes []byte
	if res.Body != nil {
		defer res.Body.Close()
		bytes, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return string(bytes), res.StatusCode, err
		}
	}

	return string(bytes), res.StatusCode, nil
}

// ApiGetBind submits an API GET request and binds the response to the
// specified interface{}.
func (c *Controller) ApiGetBind(path string, options ApiOptions, v interface{}) (int, error) {
	res, err := c.ApiRequest("GET", path, options)
	if err != nil {
		return 0, err
	}

	if res.Body == nil {
		return res.StatusCode, errors.New("Response body is empty")
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		err = json.NewDecoder(res.Body).Decode(v)
		if err != nil {
			return res.StatusCode, err
		}
	}

	return res.StatusCode, nil
}

// GetReader returns an io.Reader which will read the content of an interface{}
// as JSON encoded data. Used to create a HTTP request body from an
// interface{}.
func (c *Controller) GetReader(body interface{}) (io.Reader, error) {
	if str, ok := body.(string); ok {
		return strings.NewReader(str), nil
	}

	b, err := json.Marshal(body)
	if err != nil {
		revel.ERROR.Panicf("Failed to encode request body: %#v", body)
		return nil, err
	}

	return strings.NewReader(string(b)), nil
}

func (c *Controller) ApiPost(path string, options ApiOptions, body interface{}) (*http.Response, error) {
	reader, err := c.GetReader(body)
	if err != nil {
		return nil, err
	}

	options.Body = reader
	return c.ApiRequest("POST", path, options)
}

func (c *Controller) ApiPut(path string, options ApiOptions, body interface{}) (*http.Response, error) {
	reader, err := c.GetReader(body)
	if err != nil {
		return nil, err
	}

	options.Body = reader
	return c.ApiRequest("PUT", path, options)
}

// Bind decodes the body of a HTTP response into the specified interface{}.
func (c *Controller) Bind(res *http.Response, v interface{}) error {
	if res.Body == nil {
		return errors.New("Response body is empty")
	}
	defer res.Body.Close()

	if ctype := res.Header.Get("Content-Type"); ctype != "application/json" {
		return errors.New(fmt.Sprintf("Invalid content type: %s", ctype))
	}

	err := json.NewDecoder(res.Body).Decode(v)

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

// AuthContext returns the currently authenticated user, the user's tencancy
// and available CMDBs.
func (c *Controller) AuthContext() *AuthContext {
	options := ApiOptions{Impersonate: true}

	// Check for the auth key in the session cookie
	if !c.IsLoggedIn() {
		c.authContext = nil
		return nil
	}

	if c.authContext == nil {
		// fetch user details
		var user UserModel
		status, err := c.ApiGetBind("/users/current", options, &user)
		c.Check(err)
		if status != http.StatusOK {
			revel.ERROR.Panicf("Failed get current user from the API with: %d", status)
		}

		// fetch tenant details
		var tenant TenantModel
		status, err = c.ApiGetBind("/tenants/current", options, &tenant)
		c.Check(err)
		if status != http.StatusOK {
			revel.ERROR.Panicf("Failed get current user tenancy from the API with: %d", status)
		}

		// fetch available cmdbs
		var cmdbs []CmdbModel
		status, err = c.ApiGetBind("/cmdbs", options, &cmdbs)
		c.Check(err)
		if status != http.StatusOK {
			revel.ERROR.Panicf("Failed get a list of CMDBs from the API with: %d", status)
		}

		c.authContext = &AuthContext{
			User:   user,
			Tenant: tenant,
			Cmdbs:  cmdbs,
		}
	}

	return c.authContext
}

// IsLoggedIn checks for the presents of a session cookie and returns true if
// a valid cookie is present.
func (c *Controller) IsLoggedIn() bool {
	return c.Session["token"] != ""
}

// DestroySession clears a user's session, effectively logging them out.
func (c *Controller) DestroySession() {
	revel.TRACE.Print("Destroying user session")
	for k := range c.Session {
		delete(c.Session, k)
	}
}

// CheckLogin is an interceptor which redirects users to the login screen if
// they attempt to access a private resource without being logged in.
func (c *Controller) CheckLogin() revel.Result {
	// Check if auth token is set
	if !c.IsLoggedIn() {
		revel.TRACE.Printf("Received unauthorized request for: %s", c.Request.URL)
		// Scrub cookie
		c.DestroySession()

		// redirect to login
		c.Flash.Error("Please log in first")
		return c.Redirect(Auth.Login)
	}

	c.RenderArgs["AuthContext"] = c.AuthContext()

	return nil
}

// GetCmdb returns a struct for the requested CMDB if it is available in the
// authenticated user's tenancy. If the CMDB does not exist, nil is returned.
func (c *Controller) GetCmdb(cmdbName string) *CmdbModel {
	authContext := c.AuthContext()
	if authContext == nil {
		return nil
	}

	for _, cmdb := range authContext.Cmdbs {
		if cmdb.ShortName == cmdbName {
			c.Session["cmdb"] = cmdbName
			return &cmdb
		}
	}

	return nil
}

// GetContextCmdb returns the CMDB required for the current request context.
// The appropriate CMDB is determined in the following order of preference:
// 1. The CMDB described in the URL format /cmdbs/:cmdb (set in Session["cmdb"]
//    by ValidateRouteCmdb())
// 2. The CMDB stored in the session cookie
// 3. The first CMDB associated with the user
func (c *Controller) GetContextCmdb() *CmdbModel {
	// Get CMDB from session cookie
	cmdb := c.GetCmdb(c.Session["cmdb"])
	if cmdb != nil {
		return cmdb
	}

	// Get first available CMDB for authenticated user
	authContext := c.AuthContext()
	if authContext == nil || len(authContext.Cmdbs) == 0 {
		return nil
	}

	return &authContext.Cmdbs[0]
}

// SetSessionCmdb sets the selected CMDB cookie so all subsequent requests
// assume the specified CMDB if none is specified.
// Returns true if the CMDB exists.
func (c *Controller) SetSessionCmdb(cmdbName string) bool {
	// Do nothing if already set
	if cmdbName == c.Session["cmdb"] {
		return true
	}

	// Validate and update
	cmdb := c.GetCmdb(cmdbName)
	if cmdb != nil {
		c.Session["cmdb"] = cmdb.Name
		return true
	}

	return false
}

// Cmdb is an intercepter that ensures the existance of the CMDB required for
// the current request context.
func (c *Controller) ValidateRouteCmdb() revel.Result {
	// Get the CMDB name from the URL path
	cmdb := c.Params.Get("cmdb")
	if cmdb == "" {
		revel.ERROR.Panic("No CMDB is defined in the current route")
	}

	// Ensure the CMDB exists
	if c.GetCmdb(cmdb) == nil {
		return c.NotFound("No such CMDB was found: %s", cmdb)
	}

	// Store the CMDB in session cookie
	if !c.SetSessionCmdb(cmdb) {
		revel.ERROR.Panic("Failed to set session CMDB")
	}
	return nil
}

// AddRenderArgs is an intercepter which adds common render args to the
// controller for use in templates.
func (c *Controller) AddRenderArgs() revel.Result {
	// AppName from config file
	c.RenderArgs["AppName"], _ = revel.Config.String("app.name")

	if c.IsLoggedIn() {
		// Add current/default CMDB
		c.RenderArgs["cmdb"] = c.GetContextCmdb()
	}
	return nil
}
