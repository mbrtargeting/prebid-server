package stroeercore

import (
	"encoding/json"
	"fmt"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
	"net/http"
	"strings"
)

type StroeerCoreBidder struct {
	Url string `json:"url"`
}

type StroeerRootRequest struct {
	Id   string              `json:"id"`
	Ssat int8                `json:"ssat"`
	Bids []StroeerBidRequest `json:"bids"`
}

type StroeerBidRequest struct {
	Bid   string      `json:"bid"`
	Sid   string      `json:"sid"`
	Sizes [][2]uint64 `json:"sizes"`
}

type StroeerRootResponse struct {
	Bids []StroeerBidResponse `json:"bids"`
}

type StroeerBidResponse struct {
	BidId  string  `json:"bidId"`
	Cpm    float64 `json:"cpm"`
	Width  uint64  `json:"width"`
	Height uint64  `json:"height"`
	Ad     string  `json:"ad"`
}

func (a *StroeerCoreBidder) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errors []error
	stroeerResponse := StroeerRootResponse{}

	if err := json.Unmarshal(response.Body, &stroeerResponse); err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(len(stroeerResponse.Bids))

	for _, bid := range stroeerResponse.Bids {
		openRtbBid := openrtb.Bid{
			ImpID: bid.BidId,
			W:     bid.Width,
			H:     bid.Height,
			Price: bid.Cpm,
			AdM:   bid.Ad,
		}

		bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
			Bid:     &openRtbBid,
			BidType: openrtb_ext.BidTypeBanner,
		})
	}

	return bidderResponse, errors
}

func (b *StroeerCoreBidder) MakeRequests(internalRequest *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	errors := make([]error, 0, len(internalRequest.Imp))

	stroeerRequest := StroeerRootRequest{}
	stroeerRequest.Id = internalRequest.ID
	stroeerRequest.Ssat = 2

	stroeerRequest.Bids = []StroeerBidRequest{}

	for _, imp := range internalRequest.Imp {
		stroeerBidRequest := StroeerBidRequest{}
		stroeerBidRequest.Bid = imp.ID

		for _, format := range imp.Banner.Format {
			stroeerBidRequest.Sizes = append(stroeerBidRequest.Sizes, [2]uint64{format.W, format.H})
		}

		var bidderExt adapters.ExtImpBidder
		if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
			errors = append(errors, err)
			continue
		}

		var stroeerExt openrtb_ext.ExtImpStroeercore
		if err := json.Unmarshal(bidderExt.Bidder, &stroeerExt); err != nil {
			errors = append(errors, err)
			continue
		}

		stroeerBidRequest.Sid = stroeerExt.Sid

		stroeerRequest.Bids = append(stroeerRequest.Bids, stroeerBidRequest)
	}

	reqJSON, err := json.Marshal(stroeerRequest)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	headers.Add("User-Agent", internalRequest.Device.UA)
	headers.Add("X-Forwarded-For", internalRequest.Device.IP)

	if internalRequest.User != nil {
		userID := strings.TrimSpace(internalRequest.User.BuyerUID)
		if len(userID) > 0 {
			headers.Add("Cookie", fmt.Sprintf("uu=%s", userID))
		}
	}

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
