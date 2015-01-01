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
	"log"
	"net/http"
	"strings"
)

type Controller struct {
	*revel.Controller
	userContext *UserContext
}

func (c Controller) Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func (c Controller) ApiRequest(method string, path string, body io.Reader) (*http.Response, error) {
	// TODO: Add configurable API url
	url := fmt.Sprintf("http://localhost:3000/api/v1%s", path)
	method = strings.ToUpper(method)
	apiKey := c.Session["apiKey"]

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
		panic("An error occurred communicating with backend services")
	}

	return res, err
}

func (c Controller) ApiGet(path string) (*http.Response, error) {
	return c.ApiRequest("GET", path, nil)
}

func (c Controller) ApiGetBind(path string, v interface{}) error {
	res, err := c.ApiRequest("GET", path, nil)
	if err != nil {
		return err
	}

	if res.Body == nil {
		return errors.New("Response body is empty")
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(v)
	if err != nil {
		return err
	}

	return nil
}

func (c Controller) ApiPost(path string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Panicf("Failed to encode request body for API request to URL: %s", path)
	}

	reader := strings.NewReader(string(b))
	return c.ApiRequest("POST", path, reader)
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

func (c Controller) UserContext() *UserContext {
	if c.Session["apiKey"] == "" {
		c.userContext = nil
		return nil
	}

	if c.userContext == nil {
		// fetch user details
		var user UserModel
		err := c.ApiGetBind("/users/current", &user)
		c.Check(err)

		// fetch tenant details
		var tenant TenantModel
		err = c.ApiGetBind("/tenants/current", &tenant)
		c.Check(err)

		// fetch available cmdbs
		var cmdbs []CmdbModel
		err = c.ApiGetBind("/cmdbs", &cmdbs)
		c.Check(err)

		c.userContext = &UserContext{
			User:   user,
			Tenant: tenant,
			Cmdbs:  cmdbs,
		}
	}
	return c.userContext
}
