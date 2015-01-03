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
	firstName := c.Params.Get("firstname")
	lastName := c.Params.Get("lastname")
	email := c.Params.Get("email")
	password := c.Params.Get("password")
	password2 := c.Params.Get("password2")
	tenantCode := c.Params.Get("tenant")

	// Validate form
	c.Validation.Required(email)
	c.Validation.Required(password)
	c.Validation.Required(password == password2).Message("Passwords do not match")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Auth.Register)
	}

	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.Password = password

	// Get/Create tenant
	if tenantCode == "" {
		// Create a new tenant
		tenant.Name = email
		res, err := c.ApiPost("/tenants", &tenant)

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
		_, err = c.ApiGetBind(tenantUrl, &tenant)
		c.Check(err)

		tenantCode = tenant.Code
		user.TenantId = tenant.Id
	} else {
		// Find an existing tenant
		status, err := c.ApiGetBind(fmt.Sprintf("/tenants/%s", tenantCode), &tenant)
		c.Check(err)
		switch status {
		case http.StatusOK:
			user.TenantId = tenant.Id
		case http.StatusNotFound:
			c.Flash.Error(fmt.Sprintf("No tenant found with code: %s", tenantCode))
			return c.Redirect(Auth.Register)
		default:
			revel.ERROR.Panicf("Failed to find existing tenant with status: %d", status)
		}

		user.TenantId = tenant.Id
	}

	// Create the user
	res, err := c.ApiPost("/users", &user)
	c.Check(err)
	switch res.StatusCode {
	case http.StatusCreated:
		// Log the new user in
		return c.ValidateLogin(email, password)

	case http.StatusConflict:
		c.Flash.Error(fmt.Sprintf("An account is already registered for %s", email))
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
	res, err := c.ApiPost("/apikey", body)
	c.Check(err)

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

		// Store the auth token in the session cookie
		c.Session["token"] = apiKey

		// Now that the apiKey is set, fetch the current user details
		// TODO: Stop fetching user context to validate login
		context := c.AuthContext()
		c.Flash.Success("Welcome %s!", context.User.DisplayName())

		// Great success
		return c.Redirect(App.Index)
	default:
		// TODO: Forward unknown errors
		return c.RenderError(errors.New("Uh-oh! Basghetti Oooooh!"))
	}
}

func (c Auth) Logout() revel.Result {
	c.DestroySession()

	return c.Redirect(Auth.Login)
}
