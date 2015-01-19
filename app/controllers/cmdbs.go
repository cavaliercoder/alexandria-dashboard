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
	"github.com/revel/revel"
	"net/http"
)

type Cmdbs struct {
	Controller
}

func (c Cmdbs) Index() revel.Result {
	authContext := c.AuthContext()

	c.RenderArgs["cmdbs"] = authContext.Cmdbs

	return c.Render()
}

func (c Cmdbs) Get(cmdb string) revel.Result {
	db := c.GetCmdb(cmdb)
	if db != nil {
		c.SetSessionCmdb(cmdb)
		return c.Redirect(App.Index)
	}

	return c.NotFound("")
}

func (c Cmdbs) New() revel.Result {
	return c.Render()
}

func (c Cmdbs) ProcessNew() revel.Result {
	var cmdb CmdbModel
	cmdb.Name = c.Params.Get("name")
	cmdb.Description = c.Params.Get("description")

	// Validate params
	c.Validation.Required(cmdb.Name)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Cmdbs.New)
	}

	// Create new CMDB
	res, err := c.ApiPost(true, "/cmdbs", &cmdb)
	c.Check(err)

	// Parse the response
	switch res.StatusCode {
	case http.StatusCreated:
		c.Flash.Success("Created %s", cmdb.Name)
		c.Session["cmdb"] = cmdb.Name
		return c.Redirect(Cmdbs.Index)
	case http.StatusConflict:
		c.Flash.Error("A CMDB already exists with name '%s'", cmdb.Name)
		return c.Redirect(Cmdbs.New)
	default:
		revel.ERROR.Panicf("Failed to create new CMDB with: %s", res.Status)
	}

	return c.Redirect(Cmdbs.Index)
}
