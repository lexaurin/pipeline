// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backups

import (
	"net/http"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"

	arkAPI "github.com/banzaicloud/pipeline/internal/ark/api"
	"github.com/banzaicloud/pipeline/internal/platform/gin/correlationid"
	"github.com/banzaicloud/pipeline/src/api/ark/common"
)

// Create creates an ARK backup
func Create(c *gin.Context) {
	logger := correlationid.Logger(common.Log, c)
	logger.Info("creating backup")

	var req arkAPI.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.WrapIf(err, "could not parse request")
		common.ErrorHandler.Handle(err)
		common.ErrorResponse(c, err)
		return
	}

	err := common.GetARKService(c.Request).GetClusterBackupsService().Create(req)
	if err != nil {
		common.ErrorHandler.Handle(err)
		common.ErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, &arkAPI.CreateBackupResponse{
		Name:   req.Name,
		Status: http.StatusOK,
	})
}
