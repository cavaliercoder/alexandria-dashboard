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
)

var ApiAuthError = errors.New("API Authentication error")

type Controller struct {
	*revel.Controller
	authContext *AuthContext
}

func (c Controller) Check(err error) {
	if err != nil {
		revel.ERROR.Panic(err)
	}
}

func (c Controller) ApiRequest(impersonate bool, method string, path string, body io.Reader) (*http.Response, error) {
	// TODO: Add configurable API url
	baseUrl, ok := revel.Config.String("api.url")
	if !ok {
		panic("API URL is not set")
	}

	// Strip API version prefix
	r := regexp.MustCompile("^/api/v1")
	path = r.ReplaceAllString(path, "")

	// Build URL
	url := fmt.Sprintf("%s%s", baseUrl, path)
	method = strings.ToUpper(method)

	var apiKey string
	if impersonate {
		apiKey = c.Session["token"]
	} else {
		apiKey, ok = revel.Config.String("api.key")
		if !ok {
			revel.ERROR.Panic("API authentication key is not set")
		}
	}

	revel.TRACE.Printf("Started API Request: %s %s", method, url)

	// Create a HTTP client that does not follow redirects
	// This allows 'Location' headers to be read
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("Never follow redirects")
		},
	}

	// Create the request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Add request headers
	if method == "POST" {
		req.Header.Add("Content-type", "application/json")
	}

	if apiKey != "" {
		req.Header.Add("X-Auth-Token", apiKey)
	}

	// TODO: Update Dashboard user agent string with version info
	req.Header.Add("User-Agent", "Alexandria CMDB Dashboard")

	// Submit the request
	res, err := client.Do(req)
	if res == nil {
		revel.ERROR.Panic("An error occurred communicating with backend services")
	}

	// Validate response
	if res.StatusCode == http.StatusUnauthorized {
		return res, ApiAuthError
	}

	revel.TRACE.Printf("Finished API request with: %s", res.Status)

	return res, err
}

func (c Controller) ApiGet(impersonate bool, path string) (*http.Response, error) {
	return c.ApiRequest(impersonate, "GET", path, nil)
}

func (c Controller) ApiGetString(impersonate bool, path string) (string, int, error) {
	res, err := c.ApiRequest(impersonate, "GET", path, nil)
	if err != nil {
		return "", 0, err
	}

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

func (c Controller) BindJson(body io.Reader, v interface{}) error {
	err := json.NewDecoder(body).Decode(v)
	return err
}

func (c Controller) ApiGetBind(impersonate bool, path string, v interface{}) (int, error) {
	res, err := c.ApiRequest(impersonate, "GET", path, nil)
	if err != nil {
		return 0, err
	}

	if res.Body == nil {
		return res.StatusCode, errors.New("Response body is empty")
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		err = c.BindJson(res.Body, v)
		if err != nil {
			return res.StatusCode, err
		}
	}

	return res.StatusCode, nil
}

func (c Controller) ApiPost(impersonate bool, path string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		revel.ERROR.Panicf("Failed to encode POST request body for API request to URL: %s", path)
	}

	reader := strings.NewReader(string(b))
	return c.ApiRequest(impersonate, "POST", path, reader)
}

func (c Controller) ApiPut(impersonate bool, path string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		revel.ERROR.Panicf("Failed to encode PUT request body for API request to URL: %s", path)
	}

	reader := strings.NewReader(string(b))
	return c.ApiRequest(impersonate, "PUT", path, reader)
}

// Bind decodes the body of a HTTP response into the specified interface
func (c Controller) Bind(res *http.Response, v interface{}) error {
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

func (c Controller) AuthContext() *AuthContext {
	// Check for the auth key in the session cookie
	if !c.IsLoggedIn() {
		c.authContext = nil
		return nil
	}

	if c.authContext == nil {
		// fetch user details
		var user UserModel
		status, err := c.ApiGetBind(true, "/users/current", &user)
		c.Check(err)
		if status != http.StatusOK {
			revel.ERROR.Panicf("Failed get current user from the API with: %d", status)
		}

		// fetch tenant details
		var tenant TenantModel
		status, err = c.ApiGetBind(true, "/tenants/current", &tenant)
		c.Check(err)
		if status != http.StatusOK {
			revel.ERROR.Panicf("Failed get current user tenancy from the API with: %d", status)
		}

		// fetch available cmdbs
		var cmdbs []CmdbModel
		status, err = c.ApiGetBind(true, "/cmdbs", &cmdbs)
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

func (c Controller) IsLoggedIn() bool {
	return c.Session["token"] != ""
}

func (c Controller) DestroySession() {
	revel.TRACE.Print("Destroying user session")
	for k := range c.Session {
		delete(c.Session, k)
	}
}

// CheckLogin is an interceptor which redirects users to the login screen if
// they attempt to access a private resource without being logged in.
func (c Controller) CheckLogin() revel.Result {
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

func (c Controller) GetCmdb(cmdbName string) *CmdbModel {
	authContext := c.AuthContext()
	if authContext == nil {
		return nil
	}

	for _, cmdb := range authContext.Cmdbs {
		if cmdb.Name == cmdbName {
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
func (c Controller) GetContextCmdb() *CmdbModel {
	cmdb := c.GetCmdb(c.Session["cmdb"])
	if cmdb != nil {
		return cmdb
	}

	authContext := c.AuthContext()
	if authContext == nil || len(authContext.Cmdbs) == 0 {
		return nil
	}

	return &authContext.Cmdbs[0]
}

func (c Controller) SetSessionCmdb(cmdbName string) bool {
	if cmdbName == c.Session["cmdb"] {
		return true
	}

	cmdb := c.GetCmdb(cmdbName)
	if cmdb != nil {
		c.Session["cmdb"] = cmdb.Name
		return true
	}

	return false
}

// Cmdb is an intercepter ensures the existance of the CMDB required for the
// current request context.
// The appropriate CMDB is determined in the following order of preference:
// 1. The CMDB described in the URL format /cmdbs/:cmdb
// 2. The CMDB stored in the session cookie
// 3. The first CMDB associated with the user
func (c Controller) ValidateRouteCmdb() revel.Result {
	cmdb := c.Params.Get("cmdb")
	if cmdb == "" {
		revel.ERROR.Panic("No CMDB is defined in the current route")
	}

	if c.GetCmdb(cmdb) == nil {
		return c.NotFound("No such CMDB was found: %s", cmdb)
	}

	if !c.SetSessionCmdb(cmdb) {
		revel.ERROR.Panic("Failed to set session CMDB")
	}
	return nil
}

// AddRenderArgs is an intercepter which adds common render args to the
// controller for use in templates.
func (c Controller) AddRenderArgs() revel.Result {
	// AppName from config file
	c.RenderArgs["AppName"], _ = revel.Config.String("app.name")

	if c.IsLoggedIn() {
		// Add current/default CMDB
		c.RenderArgs["cmdb"] = c.GetContextCmdb()
	}
	return nil
}
