// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted to handle a subset of Pull Request webooks from Azure Devops
// Azure Devops Events docs: https://docs.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops

package azuredevops

// ItemContent describes an item
type ItemContent struct {
	Content     *string          `json:"content,omitempty"`
	ContentType *ItemContentType `json:"contentType,omitempty"`
}

// ItemContentType describes an item content type
type ItemContentType struct {
	Base64Encoded *string `json:"base64Encoded,omitempty"`
	RawText       *string `json:"rawText,omitempty"`
}

// Link A single item in a collection of Links.
type Link struct {
	Href *string `json:"href,omitempty"`
}

// ResourceContainers provides information related to the Resources in a payload
type ResourceContainers struct {
	Collection *ResourceRef `json:"text,omitempty"`
	Account    *ResourceRef `json:"html,omitempty"`
	Project    *ResourceRef `json:"markdown,omitempty"`
}

// ResourceRef Describes properties to identify a resource
type ResourceRef struct {
	ID      *string `json:"id,omitempty"`
	BaseURL *string `json:"baseUrl,omitempty"`
	URL     *string `json:"url,omitempty"`
}

// WebAPITagDefinition The representation of a tag definition which is sent across
// the wire.
type WebAPITagDefinition struct {
	Active *bool   `json:"active,omitempty"`
	ID     *string `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	URL    *string `json:"url,omitempty"`
}
