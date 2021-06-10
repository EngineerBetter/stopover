package concourse

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse/internal"
	"github.com/tedsuo/rata"
)

func (team *team) CheckResourceType(pipelineRef atc.PipelineRef, resourceTypeName string, version atc.Version) (atc.Build, bool, error) {

	params := rata.Params{
		"pipeline_name":      pipelineRef.Name,
		"resource_type_name": resourceTypeName,
		"team_name":          team.Name(),
	}

	var build atc.Build

	jsonBytes, err := json.Marshal(atc.CheckRequestBody{From: version})
	if err != nil {
		return build, false, err
	}

	err = team.connection.Send(internal.Request{
		RequestName: atc.CheckResourceType,
		Params:      params,
		Query:       pipelineRef.QueryParams(),
		Body:        bytes.NewBuffer(jsonBytes),
		Header:      http.Header{"Content-Type": []string{"application/json"}},
	}, &internal.Response{
		Result: &build,
	})

	switch e := err.(type) {
	case nil:
		return build, true, nil
	case internal.ResourceNotFoundError:
		return build, false, nil
	case internal.UnexpectedResponseError:
		if e.StatusCode == http.StatusInternalServerError {
			return build, false, GenericError{e.Body}
		} else {
			return build, false, err
		}
	default:
		return build, false, err
	}
}
