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
	"fmt"
	"github.com/revel/revel"
	"net/http"
)

type CITypes struct {
	Controller
}

func (c CITypes) Index(id string) revel.Result {
	cmdb := c.GetContextCmdb()

	// Get CI Types
	var citypes []CITypeModel
	status, err := c.ApiGetBind(true, fmt.Sprintf("/cmdbs/%s/citypes", cmdb.ShortName), &citypes)
	c.Check(err)

	if status != http.StatusOK {
		revel.ERROR.Panicf("Failed to retrieve CI Types for database %s with: %d", cmdb.Name, status)
	}

	c.RenderArgs["citypes"] = citypes

	// Get selected CI Type
	if id == "" {
		if 0 < len(citypes) {
			c.RenderArgs["citype"] = &citypes[0]
		}
	} else {
		found := false
		for _, citype := range citypes {
			if citype.ShortName == id {
				c.RenderArgs["citype"] = &citype
				found = true
				break
			}
		}

		if !found {
			return c.NotFound("No such CI Type: %s", id)
		}
	}
	return c.Render()
}

func (c CITypes) ProcessNew() revel.Result {
	var citype CITypeModel
	citype.Name = c.Params.Get("name")
	citype.Description = c.Params.Get("description")

	// Validate params
	c.Validation.Required(citype.Name)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(CITypes.Index)
	}

	// Create the CI Type
	cmdb := c.GetContextCmdb()
	res, err := c.ApiPost(true, fmt.Sprintf("/cmdbs/%s/citypes", cmdb.ShortName), &citype)
	c.Check(err)
	switch res.StatusCode {
	case http.StatusCreated:
		c.Flash.Success("Created %s", citype.Name)
		return c.Redirect("/cmdb/%s/citypes/%s", cmdb.ShortName, citype.ShortName)

	case http.StatusConflict:
		c.Flash.Error("CI type '%s' already exists", citype.Name)
		return c.Redirect("/cmdb/%s/citypes", cmdb.Name)
	default:
		revel.ERROR.Panicf("Failed to create CI Type with: %s", res.Status)
	}

	return nil
}
