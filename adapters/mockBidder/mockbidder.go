package mockBidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/prebid/openrtb/v20/openrtb2"

	"github.com/prebid/prebid-server/v4/adapters"
	"github.com/prebid/prebid-server/v4/config"
	"github.com/prebid/prebid-server/v4/errortypes"
	"github.com/prebid/prebid-server/v4/openrtb_ext"
	"github.com/prebid/prebid-server/v4/util/jsonutil"
)

type adapter struct {
	URL    string `json:"url"`
	Server config.Server
}

type response struct {
	Bids []bidResponse `json:"bids"`
}

type bidResponse struct {
	ID     string          `json:"id"`
	BidID  string          `json:"bidId"`
	CPM    float64         `json:"cpm"`
	Width  int64           `json:"width"`
	Height int64           `json:"height"`
	Ad     string          `json:"ad"`
	CrID   string          `json:"crid"`
	Mtype  string          `json:"mtype"`
	DSA    json.RawMessage `json:"dsa"`
}

type bidExt struct {
	DSA json.RawMessage `json:"dsa,omitempty"`
	Abc string          `json:"abc,omitempty"`
}

func (b *bidResponse) resolveMediaType() (mt openrtb2.MarkupType, bt openrtb_ext.BidType, err error) {
	switch b.Mtype {
	case "banner":
		return openrtb2.MarkupBanner, openrtb_ext.BidTypeBanner, nil
	case "video":
		return openrtb2.MarkupVideo, openrtb_ext.BidTypeVideo, nil
	default:
		return mt, bt, fmt.Errorf("unable to determine media type for bid with id \"%s\"", b.BidID)
	}
}

func (a *adapter) MakeBids(request *openrtb2.BidRequest, _ *adapters.RequestData, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errs []error

	switch responseData.StatusCode {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusBadRequest:
		return nil, []error{&errortypes.BadInput{
			Message: "unexpected status code: " + strconv.Itoa(responseData.StatusCode),
		}}
	case http.StatusOK:
		break
	default:
		return nil, []error{&errortypes.BadServerResponse{
			Message: "unexpected status code: " + strconv.Itoa(responseData.StatusCode),
		}}
	}

	var bidResponse openrtb2.BidResponse
	err := jsonutil.Unmarshal(responseData.Body, &bidResponse)
	if err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: err.Error(),
		}}
	}

	response := adapters.NewBidderResponseWithBidsCapacity(len(request.Imp))

	for _, seatBid := range bidResponse.SeatBid {
		for i := range seatBid.Bid {
			response.Bids = append(response.Bids, &adapters.TypedBid{
				Bid:     &seatBid.Bid[i],
				BidType: getMediaTypeForImp(seatBid.Bid[i].ImpID, request.Imp),
			})
		}
	}

	return response, errs
}

func getMediaTypeForImp(impID string, imps []openrtb2.Imp) openrtb_ext.BidType {
	for _, imp := range imps {
		if imp.ID == impID {
			if imp.Banner != nil {
				return openrtb_ext.BidTypeBanner
			} else if imp.Video != nil {
				return openrtb_ext.BidTypeVideo
			}
		}
	}
	return openrtb_ext.BidTypeBanner
}

func (a *adapter) MakeRequests(bidRequest *openrtb2.BidRequest, extraRequestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var errors []error

	for idx := range bidRequest.Imp {
		imp := &bidRequest.Imp[idx]
		var bidderExt adapters.ExtImpBidder
		if err := jsonutil.Unmarshal(imp.Ext, &bidderExt); err != nil {
			errors = append(errors, err)
			continue
		}

		var stroeerExt openrtb_ext.ExtImpStroeerCore
		if err := jsonutil.Unmarshal(bidderExt.Bidder, &stroeerExt); err != nil {
			errors = append(errors, err)
			continue
		}

		imp.TagID = stroeerExt.Sid
	}

	reqJSON, err := json.Marshal(bidRequest)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.URL,
		Body:    reqJSON,
		Headers: headers,
		ImpIDs:  openrtb_ext.GetImpIDs(bidRequest.Imp),
	}}, errors
}

// Builder builds a new instance of the MockBidder adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	bidder := &adapter{
		URL: config.Endpoint,
	}
	return bidder, nil
}
