/*
 * Copyright (c) 2017, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package httpbuilder

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/megaease/easegress/pkg/context"
	"github.com/megaease/easegress/pkg/filters"
	"github.com/megaease/easegress/pkg/logger"
	"github.com/megaease/easegress/pkg/protocols/httpprot"
	"github.com/megaease/easegress/pkg/util/yamltool"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMain(m *testing.M) {
	logger.InitNop()
	code := m.Run()
	os.Exit(code)
}

func getRequestBuilder(spec *HTTPRequestBuilderSpec) *HTTPRequestBuilder {
	rb := &HTTPRequestBuilder{spec: spec}
	rb.Init()
	return rb
}

func TestMethod(t *testing.T) {
	assert := assert.New(t)

	// get method from request
	// directly set body
	yml := `template: |
  method: {{ .requests.request1.Method }}
  url: /
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com?field1=value1&field2=value2", nil)
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request).Std()
		assert.Equal(http.MethodDelete, testReq.Method)
	}

	// set method directly
	yml = `template: |
  method: get
  url: /
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com?field1=value1&field2=value2", nil)
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request).Std()
		assert.Equal(http.MethodGet, testReq.Method)
	}

	// invalid method
	yml = `template: |
  method: what
  url: /
`
	{

		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)
		res := rb.Handle(ctx)
		assert.NotEmpty(res)
	}
}

func TestURL(t *testing.T) {
	assert := assert.New(t)

	// get url from request
	yml := `template: |
  method: Delete
  url:  http://www.facebook.com?field1={{index .requests.request1.URL.Query.field2 0}}
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com?field1=value1&field2=value2", nil)
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request).Std()
		assert.Equal(http.MethodDelete, testReq.Method)
		assert.Equal("http://www.facebook.com?field1=value2", testReq.URL.String())
	}

	// set url directly
	yml = `template: |
  method: Put
  url:  http://www.facebook.com
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com?field1=value1&field2=value2", nil)
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request).Std()
		assert.Equal(http.MethodPut, testReq.Method)
		assert.Equal("http://www.facebook.com", testReq.URL.String())
	}
}

func TestRequestHeader(t *testing.T) {
	assert := assert.New(t)

	// get header from request and response
	yml := `template: |
  method: Delete
  url:  http://www.facebook.com
  headers:
    "X-Request": [{{index (index .requests.request1.Header "X-Request") 0}}]
    "X-Response": [{{index (index .responses.response1.Header "X-Response") 0}}]
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com?field1=value1&field2=value2", nil)
		assert.Nil(err)
		req1.Header.Add("X-Request", "from-request1")
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		resp1 := &http.Response{}
		resp1.Header = http.Header{}
		resp1.Header.Add("X-Response", "from-response1")
		httpresp1, err := httpprot.NewResponse(resp1)
		assert.Nil(err)
		ctx.SetResponse("response1", httpresp1)

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request).Std()
		assert.Equal(http.MethodDelete, testReq.Method)
		assert.Equal("http://www.facebook.com", testReq.URL.String())
		assert.Equal("from-request1", testReq.Header.Get("X-Request"))
		assert.Equal("from-response1", testReq.Header.Get("X-Response"))
	}
}

func TestRequestBody(t *testing.T) {
	assert := assert.New(t)

	// directly set body
	yml := `template: |
  method: Delete
  url:  http://www.facebook.com
  body: body
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)
		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com", nil)
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request)
		data, err := io.ReadAll(testReq.GetPayload())
		assert.Nil(err)
		assert.Equal("body", string(data))
	}

	// set body by using other body
	yml = `template: |
  method: Delete
  url:  http://www.facebook.com
  body: body {{ .requests.request1.Body }}
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com", strings.NewReader("123"))
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request)
		data, err := io.ReadAll(testReq.GetPayload())
		assert.Nil(err)
		assert.Equal("body 123", string(data))
	}

	// set body by using json map
	yml = `template: |
  method: Delete
  url:  http://www.facebook.com
  body: body {{ .requests.request1.JSONBody.field1 }} {{ .requests.request1.JSONBody.field2 }}
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com", strings.NewReader(`{"field1":"value1", "field2": "value2"}`))
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request)
		data, err := io.ReadAll(testReq.GetPayload())
		assert.Nil(err)
		assert.Equal("body value1 value2", string(data))
	}

	// set body by using yaml map
	yml = `template: |
  method: Delete
  url:  http://www.facebook.com
  body: body {{ .requests.request1.YAMLBody.field1 }} {{ .requests.request1.YAMLBody.field2 }}
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)

		req1, err := http.NewRequest(http.MethodDelete, "http://www.google.com", strings.NewReader(`
field1: value1
field2: value2
`))
		assert.Nil(err)
		setRequest(t, ctx, "request1", req1)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request)
		data, err := io.ReadAll(testReq.GetPayload())
		assert.Nil(err)
		assert.Equal("body value1 value2", string(data))
	}

	// use default method
	yml = `template: |
  url:  http://www.facebook.com
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request)
		assert.Equal(http.MethodGet, testReq.Std().Method)
	}

	// use default url
	yml = `template: |
  method: delete 
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.Empty(res)
		testReq := ctx.GetRequest("test").(*httpprot.Request)
		assert.Equal(http.MethodDelete, testReq.Std().Method)
		assert.Equal("/", testReq.Std().URL.String())
	}

	// build request failed
	yml = `template: |
  url: http://192.168.0.%31:8080/
`
	{
		spec := &HTTPRequestBuilderSpec{}
		yaml.Unmarshal([]byte(yml), spec)
		rb := getRequestBuilder(spec)
		defer rb.Close()

		ctx := context.New(nil)
		ctx.UseNamespace("test")

		res := rb.Handle(ctx)
		assert.NotEmpty(res)
	}
}

func TestHTTPRequestBuilder(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(&HTTPRequestBuilderSpec{}, httpRequestBuilderKind.DefaultSpec())
	yamlStr := `
name: requestBuilder 
kind: HTTPRequestBuilder
template: |
  method: Delete
`
	rawSpec := map[string]interface{}{}
	yamltool.Unmarshal([]byte(yamlStr), &rawSpec)
	spec, err := filters.NewSpec(nil, "pipeline1", rawSpec)
	assert.Nil(err)
	requestBuilder := httpRequestBuilderKind.CreateInstance(spec).(*HTTPRequestBuilder)
	assert.Equal("requestBuilder", requestBuilder.Name())
	assert.Equal(httpRequestBuilderKind, requestBuilder.Kind())
	assert.Equal(spec, requestBuilder.Spec())
	requestBuilder.Init()

	newRequestBuilder := httpRequestBuilderKind.CreateInstance(spec)
	newRequestBuilder.Inherit(requestBuilder)
	assert.Nil(newRequestBuilder.Status())
}