package stroeercore

import (
	"encoding/json"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"net/http"
)

type StroeerCoreBidder struct {
	url string
}

func (a *StroeerCoreBidder) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	bidderResponse := adapters.NewBidderResponse()
	var errors []error
	return bidderResponse, errors
}

func (b *StroeerCoreBidder) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	errors := make([]error, 0, len(request.Imp))

	reqJSON, err := json.Marshal(request)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	headers := http.Header{}

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     b.url,
		Body:    reqJSON,
		Headers: headers,
	}}, errors
}

func NewStroeerCoreBidder(endpoint string) *StroeerCoreBidder {
	return &StroeerCoreBidder{url: endpoint}
}
