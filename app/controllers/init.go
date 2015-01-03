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

import "github.com/revel/revel"

func init() {
	// Log all requests in trace mode
	revel.InterceptFunc(func(c *revel.Controller) revel.Result {
		revel.TRACE.Printf("Starting: %s", c.Request.URL.String())
		return nil
	}, revel.BEFORE, Controller{})

	// Enforce authentication for private controllers
	revel.InterceptMethod(App.CheckLogin, revel.BEFORE)
	revel.InterceptMethod(Cmdbs.CheckLogin, revel.BEFORE)
	revel.InterceptMethod(CITypes.CheckLogin, revel.BEFORE)

	// Validate CMDB URL params for CMDB related routes
	revel.InterceptMethod(CITypes.ValidateRouteCmdb, revel.BEFORE)

	// Add common RenderArgs such as Application Name
	revel.InterceptMethod(Controller.AddRenderArgs, revel.BEFORE)
}
