package stroeercore

import (
	"github.com/prebid/prebid-server/adapters/adapterstest"
	"testing"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "stroeercoretest", NewStroeerCoreBidder("http://localhost:8361/xyz"))
}
