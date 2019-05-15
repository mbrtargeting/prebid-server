package stroeercore

import (
	"encoding/json"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
	"net/http"
)

type StroeerCoreBidder struct {
	Url string `json:"url"`
}

type BidderRootRequest struct {
	Id   string             `json:"id"`
	Ssat int8               `json:"ssat"`
	Bids []BidderBidRequest `json:"bids"`
}

type BidderBidRequest struct {
	Bid   string      `json:"bid"`
	Sid   string      `json:"sid"`
	Sizes [][2]uint64 `json:"sizes"`
}

type BidderRootResponse struct {
	Bids []BidderBidResponse ``
}

type BidderBidResponse struct {
	bidId  string
	cpm    float64
	width  uint64
	height uint64
	ad     string
}

func (a *StroeerCoreBidder) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	bidderResponse := adapters.NewBidderResponse()
	var errors []error
	return bidderResponse, errors
}

func (b *StroeerCoreBidder) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	errors := make([]error, 0, len(request.Imp))

	bidderRequest := BidderRootRequest{}
	bidderRequest.Id = request.ID
	bidderRequest.Ssat = 2

	bidderRequest.Bids = []BidderBidRequest{}

	imp := request.Imp[0]

	bidderBidRequest := BidderBidRequest{}
	bidderBidRequest.Bid = imp.ID

	for _, format := range imp.Banner.Format {
		bidderBidRequest.Sizes = append(bidderBidRequest.Sizes, [2]uint64{format.W, format.H})
	}

	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	var stroeerExt openrtb_ext.ExtImpStroeercore
	if err := json.Unmarshal(bidderExt.Bidder, &stroeerExt); err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	bidderBidRequest.Sid = stroeerExt.Sid

	bidderRequest.Bids = append(bidderRequest.Bids, bidderBidRequest)

	reqJSON, err := json.Marshal(bidderRequest)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	headers := http.Header{}

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     b.Url,
		Body:    reqJSON,
		Headers: headers,
	}}, errors
}

func NewStroeerCoreBidder(endpoint string) *StroeerCoreBidder {
	return &StroeerCoreBidder{Url: endpoint}
}
