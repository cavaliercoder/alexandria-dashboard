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
	"fmt"
	"github.com/revel/revel"
	"net/http"
	"regexp"
)

type CITypes struct {
	Controller
}

func (c CITypes) Index() revel.Result {
	cmdb := c.GetContextCmdb()
	options := ApiOptions{
		Impersonate: true,
		Selector: map[string]interface{}{
			"name":        1,
			"shortname":   1,
			"description": 1,
		},
	}

	// Get CI Types
	var citypes []CITypeModel
	status, err := c.ApiGetBind(fmt.Sprintf("/cmdbs/%s/citypes", cmdb.ShortName), options, &citypes)
	c.Check(err)

	if status != http.StatusOK {
		revel.ERROR.Panicf("Failed to retrieve CI Types for database %s with: %d", cmdb.Name, status)
	}

	c.RenderArgs["citypes"] = citypes

	return c.Render()
}

func (c CITypes) Edit(id string) revel.Result {
	cmdb := c.GetContextCmdb()
	options := ApiOptions{Impersonate: true}

	// Get CI Types
	var citype CITypeModel
	status, err := c.ApiGetBind(fmt.Sprintf("/cmdbs/%s/citypes/%s", cmdb.ShortName, id), options, &citype)
	c.Check(err)

	switch status {
	case http.StatusOK:
		// Do nothing

	case http.StatusNotFound:
		c.NotFound("No such CI Type: %s", id)
	}

	// Store the CI type for rendering
	c.RenderArgs["citype"] = citype

	// Store raw JSON version for javascript
	bytes, err := json.Marshal(citype)
	if err != nil {
		revel.ERROR.Panicf("Failed to marshall interface to JSON with: %s", err)
	}
	c.RenderArgs["citypeJson"] = string(bytes)

	return c.Render()
}

func (c CITypes) Add() revel.Result {
	options := ApiOptions{Impersonate: true}

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
	res, err := c.ApiPost(fmt.Sprintf("/cmdbs/%s/citypes", cmdb.ShortName), options, &citype)
	c.Check(err)
	switch res.StatusCode {
	case http.StatusCreated:
		// Get the CI Type short name from the returned location header
		location := res.Header.Get("Location")
		re := regexp.MustCompile("[^/]*$")
		shortname := re.FindString(location)

		c.Flash.Success("Created %s", citype.Name)
		return c.Redirect("/cmdb/%s/citypes/%s", cmdb.ShortName, shortname)

	case http.StatusConflict:
		c.Flash.Error("CI type '%s' already exists", citype.Name)
		return c.Redirect("/cmdb/%s/citypes", cmdb.Name)

	default:
		revel.ERROR.Panicf("Failed to create CI Type with: %s", res.Status)
	}

	return nil
}

func (c CITypes) Update(cmdb string, id string, data string) revel.Result {
	options := ApiOptions{Impersonate: true}

	// Validate request
	c.Validation.Required(cmdb)
	c.Validation.Required(id)
	c.Validation.Required(data)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(CITypes.Index)
	}

	// Get the old type
	var oldUri = fmt.Sprintf("/cmdbs/%s/citypes/%s", cmdb, id)
	var original CITypeModel
	status, err := c.ApiGetBind(oldUri, options, &original)
	c.Check(err)
	if status != http.StatusOK {
		c.Flash.Error("Failed to retrieve original CI Type: %s", id)
		return c.Redirect("/cmdb/%s/citypes/%s", cmdb, id)
	}

	// Send the update data to the API
	res, err := c.ApiPut(fmt.Sprintf("/cmdbs/%s/citypes/%s", cmdb, id), options, data)
	c.Check(err)

	// Compute URL of updated resource
	var newUri = oldUri
	switch res.StatusCode {
	case http.StatusNoContent:
		// Do nothing
	case http.StatusMovedPermanently:
		newUri = res.Header.Get("Location")
	default:
		revel.ERROR.Panicf("Failed to update resource with: %s", res.Status)
	}

	// Get the new type
	var updated CITypeModel
	status, err = c.ApiGetBind(newUri, options, &updated)
	c.Check(err)
	if status != http.StatusOK {
		revel.ERROR.Panicf("Failed to fetch updated resource with: %d", status)
	}

	// Redirect to the updated resource
	if updated.ShortName == original.ShortName {
		c.Flash.Success("Updated %s", updated.Name)
		return c.Redirect("/cmdb/%s/citypes/%s", cmdb, original.ShortName)
	} else {
		c.Flash.Success("Updated %s (Renamed to %s)", original.Name, updated.Name)
		return c.Redirect("/cmdb/%s/citypes/%s", cmdb, updated.ShortName)
	}
}
