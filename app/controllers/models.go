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

type AuthContext struct {
	User   UserModel
	Tenant TenantModel
	Cmdbs  []CmdbModel
}

type UserModel struct {
	TenantCode string `json:"tenantCode"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

func (c *UserModel) DisplayName() string {
	if c.FirstName == "" {
		return c.Email
	} else {
		return c.FirstName
	}
}

type TenantModel struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CmdbModel struct {
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description"`
}

type CITypeModel struct {
	Name        string                 `json:"name"`
	ShortName   string                 `json:"shortName"`
	Description string                 `json:"description"`
	InheritFrom string                 `json:"inheritFrom"`
	Attributes  []CITypeAttributeModel `json:"attributes"`
}

type CITypeAttributeModel struct {
	Name        string                 `json:"name"`
	ShortName   string                 `json:"shortName"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Children    []CITypeAttributeModel `json:"children"`

	// Common Options
	Required bool `json:"required,omitempty"`
	IsArray  bool `json:"isArray,omitempty"`
	MinCount int  `json:"minCount,omitempty"`
	MaxCount int  `json:"maxCount,omitempty"`

	// Group options
	Singular string `json:"singular,omitempty"`

	// String options
	MinLength uint     `json:"minLength,omitempty"`
	MaxLength uint     `json:"maxLength,omitempty"`
	Filters   []string `json:"filters,omitempty"`

	// Number options
	Units    string `json:"units,omitempty"`
	MinValue int    `json:"minValue,omitempty"`
	MaxValue int    `json:"maxValue,omitempty"`
	Signed   bool   `json:"signed,omitempty"`
	Float    bool   `json:"float,omitempty"`
}
