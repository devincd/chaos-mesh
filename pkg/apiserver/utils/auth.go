// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authorizationv1 "k8s.io/api/authorization/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func AuthRequired(c *gin.Context) {
	if mockResult := mock.On("MockAuthRequired"); mockResult != nil {
		c.Next()
		return
	}

	authCli, err := clientpool.ExtractTokenAndGetAuthClient(c.Request.Header)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	namespace := c.Query("namespace")
	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "list",
				Group:     "chaos-mesh.org",
				Resource:  "*",
			},
		},
	}

	response, err := authCli.SelfSubjectAccessReviews().Create(sar)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	if !response.Status.Allowed {
		if len(namespace) == 0 {
			c.AbortWithError(http.StatusUnauthorized, ErrNoClusterPrivilege.New("can't list chaos experiments in the cluster"))
		} else {
			c.AbortWithError(http.StatusUnauthorized, ErrNoNamespacePrivilege.New("can't list chaos experiments in namespace %s", namespace))
		}
		return
	}

	c.Next()
}
