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
	"errors"
	"github.com/cavaliercoder/alexandria-dashboard/app"
	"github.com/revel/revel"
	"net/http"
)

type App struct {
	Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Login() revel.Result {
	return c.Render()
}

func (c App) ValidateLogin(username string, password string) revel.Result {
	// TODO: Antiforgery token
	var err error

	// Validate form
	if username == "" {
		c.Flash.Error("Please specify a valid username")
		return c.Redirect(App.Login)
	}

	if password == "" {
		c.Flash.Error("Please specify a valid password")
		return c.Redirect(App.Login)
	}

	// Request API Key using username and password
	//body := strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password))
	//res, err := c.ApiRequest("POST", "/apikey", body)
	body := map[string]string{
		"username": username,
		"password": password,
	}
	res, err := c.ApiPost("/apikey", body)
	c.Check(err)

	// Parse the response
	switch res.StatusCode {
	case http.StatusUnauthorized:
		c.Flash.Error("The credentials you provided do not appear valid")
		return c.Redirect(App.Login)

	case http.StatusOK:
		// Parse the apiKey
		result := make(map[string]string)
		err = c.Bind(res, &result)
		c.Check(err)

		apiKey := result["apiKey"]
		c.Session["apiKey"] = apiKey

		// Now that the apiKey is set, fetch the current user details
		var user app.User
		res, err = c.ApiRequest("GET", "/users/current", nil)
		c.Check(err)

		err = c.Bind(res, &user)
		c.Check(err)
		c.Flash.Success("Welcome %s!", user.FirstName)

		// and fetch the tenant
		var tenant app.Tenant
		res, err = c.ApiRequest("GET", "/tenants/current", nil)
		c.Check(err)

		err = c.Bind(res, &tenant)
		c.Check(err)

		// Great success
		return c.Redirect(App.Index)
	default:
		// TODO: Forward unknown errors
		return c.RenderError(errors.New("Uh-oh! Basghetti Oooooh!"))
	}
}

func (c App) Logout() revel.Result {
	c.Session["apiKey"] = ""

	return c.Redirect(App.Login)
}
