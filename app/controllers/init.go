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
	// Enforce authentication for private controllers
	revel.InterceptMethod(App.CheckLogin, revel.BEFORE)

	// Add common RenderArgs such as Application Name
	revel.InterceptFunc(AddRenderArgs, revel.BEFORE, Controller{})

}

// InitRenderArgs is an intercepter which adds common render args to the
// controller for use in templates.
func AddRenderArgs(c *revel.Controller) revel.Result {
	// AppName from config file
	c.RenderArgs["AppName"], _ = revel.Config.String("app.name")
	return nil
}
