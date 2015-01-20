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
	"fmt"
	"github.com/revel/revel"
	"net/http"
)

type Auth struct {
	Controller
}

func (c Auth) Login() revel.Result {
	return c.Render()
}

func (c Auth) Register() revel.Result {
	return c.Render()
}

func (c Auth) ProcessRegistration() revel.Result {

	var tenant TenantModel
	var user UserModel

	// Destroy any existing session
	c.DestroySession()

	// Get fields
	user.FirstName = c.Params.Get("firstname")
	user.LastName = c.Params.Get("lastname")
	user.Email = c.Params.Get("email")
	user.Password = c.Params.Get("password")
	password2 := c.Params.Get("password2")
	user.TenantCode = c.Params.Get("tenant")

	// Validate form
	c.Validation.Required(user.Email)
	c.Validation.Required(user.Password)
	c.Validation.Required(user.Password == password2).Message("Passwords do not match")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Auth.Register)
	}

	// Get/Create tenant
	if user.TenantCode == "" {
		// Create a new tenant
		tenant.Name = user.Email
		res, err := c.ApiPost(false, "/tenants", &tenant)

		if err != nil {
			revel.ERROR.Panicf("Failed to create new tenant with: %s", err)
		} else if res.StatusCode != http.StatusCreated {
			revel.ERROR.Panicf("Failed to create new tenant with: %s", res.Status)
		}

		tenantUrl := res.Header.Get("Location")
		if tenantUrl == "" {
			revel.ERROR.Panicf("No location header returned for new tenant registration.")
		}

		// Fetch the new tenant
		_, err = c.ApiGetBind(false, tenantUrl, &tenant)
		c.Check(err)

		user.TenantCode = tenant.Code
	} else {
		// Find an existing tenant
		status, err := c.ApiGetBind(false, fmt.Sprintf("/tenants/%s", user.TenantCode), &tenant)
		c.Check(err)
		switch status {
		case http.StatusOK:
			user.TenantCode = tenant.Code
		case http.StatusNotFound:
			c.Flash.Error(fmt.Sprintf("No tenant found with code: %s", user.TenantCode))
			return c.Redirect(Auth.Register)
		default:
			revel.ERROR.Panicf("Failed to find existing tenant with status: %d", status)
		}
	}

	// Create the user
	res, err := c.ApiPost(false, "/users", &user)
	c.Check(err)
	switch res.StatusCode {
	case http.StatusCreated:
		// Log the new user in
		return c.ValidateLogin(user.Email, user.Password)

	case http.StatusConflict:
		c.Flash.Error(fmt.Sprintf("An account is already registered for %s", user.Email))
		return c.Redirect(Auth.Register)

	default:
		revel.ERROR.Panicf("Failed to create user with: %s", res.Status)
	}

	return nil
}

func (c Auth) ValidateLogin(username string, password string) revel.Result {
	// TODO: Antiforgery token
	var err error

	// Validate form
	c.Validation.Required(username)
	c.Validation.Required(password)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Auth.Login)
	}

	// Request API Key using username and password
	body := map[string]string{
		"username": username,
		"password": password,
	}
	res, err := c.ApiPost(false, "/apikey", body)
	if err != ApiAuthError {
		c.Check(err)
	}

	// Parse the response
	switch res.StatusCode {
	case http.StatusUnauthorized:
		c.Flash.Error("The credentials you provided do not appear valid")
		return c.Redirect(Auth.Login)

	case http.StatusOK:
		// Parse the apiKey from the API response
		result := make(map[string]string)
		err = c.Bind(res, &result)
		c.Check(err)
		apiKey := result["apiKey"]

		// Reset any existing session
		c.DestroySession()

		// Store the auth token in the session cookie
		c.Session["token"] = apiKey

		// Now that the apiKey is set, fetch the current user details
		// TODO: Stop fetching user context to validate login
		context := c.AuthContext()
		c.Flash.Success("Welcome %s!", context.User.DisplayName())

		// Great success. Where should we redirect to?
		if len(context.Cmdbs) == 0 {
			return c.Redirect(Cmdbs.New)
		} else {
			return c.Redirect(App.Index)
		}
	default:
		// TODO: Forward unknown errors
		return c.RenderError(errors.New("Uh-oh! Basghetti Oooooh!"))
	}
}

func (c Auth) Logout() revel.Result {
	c.DestroySession()

	return c.Redirect(Auth.Login)
}
