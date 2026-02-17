import { registerBidder } from './prebid/src/adapters/bidderFactory.js';
import { BANNER, NATIVE, VIDEO } from './prebid/src/mediaTypes.js';

const BIDDER_CODE = 'taskirx';
const ENDPOINT_URL = 'http://localhost:8080/bid'; // Local dev engine

export const spec = {
  code: BIDDER_CODE,
  supportedMediaTypes: [BANNER, NATIVE, VIDEO],

  /**
   * Determines whether or not the given bid request is valid.
   *
   * @param {object} bid The bid params to validate.
   * @return boolean True if this is a valid bid, and false otherwise.
   */
  isBidRequestValid: function(bid) {
    if (!bid.params.placementId) {
      return false; // must have placementId
    }
    return true;
  },

  /**
   * Make a server request from the list of BidRequests.
   *
   * @param {Array} validBidRequests A non-empty list of valid bid requests that should be sent to the Server.
   * @param {object} bidderRequest
   * @return ServerRequest Info describing the request to the server.
   */
  buildRequests: function(validBidRequests, bidderRequest) {
    const requests = validBidRequests.map(bid => {
      const payload = {
        id: bidderRequest.bidderRequestId,
        timestamp: new Date().toISOString(),
        publisher_id: bid.params.publisherId || 'unknown-publisher',
        ad_slot: {
          id: bid.adUnitCode,
          dimensions: bid.mediaTypes.banner ? bid.mediaTypes.banner.sizes[0] : [0,0], // simplify for now
          position: "unknown",
          formats: Object.keys(bid.mediaTypes) // ["banner", "native", "video"]
        },
        user: {
          // In real world, we would sync user IDs here
          id: 'user-' + Math.random().toString(36).substring(7),
          // Prebid passes First-Party Data here if available
        },
        device: {
          type: "desktop", // simplified, should detect
          ua: navigator.userAgent,
          // geo data would come from specialized module or IP lookup on server
          geo: { lat: 0, lon: 0 } // placeholder
        },
        context: {
          referer: bidderRequest.refererInfo.referer,
          page_url: bidderRequest.refererInfo.page
        }
      };

      return {
        method: 'POST',
        url: ENDPOINT_URL,
        data: JSON.stringify(payload)
      };
    });
    return requests;
  },

  /**
   * Unpack the response from the server into a list of bids.
   *
   * @param {object} serverResponse A successful response from the server.
   * @return {Array} An array of bids which were nested inside the server.
   */
  interpretResponse: function(serverResponse, bidRequest) {
    const serverBody = serverResponse.body;
    const bidResponses = [];

    if (!serverBody || serverBody.bid_price === 0) {
      return []; // no bid
    }

    const bid = {
      requestId: bidRequest.data ? JSON.parse(bidRequest.data).id : bidRequest.bidderRequestId, // matching request ID
      cpm: serverBody.bid_price,
      width: 300, // logic to parse from creative
      height: 250,
      creativeId: serverBody.creative_url,
      currency: 'USD',
      netRevenue: true,
      ttl: serverBody.ttl || 300,
      meta: {
        advertiserDomains: [] // optional
      },
      ad: `<a href="${serverBody.click_url}" target="_blank"><img src="${serverBody.creative_url}" /></a><img src="${serverBody.impression_url}" style="display:none" />`
    };

    bidResponses.push(bid);
    return bidResponses;
  }
};

registerBidder(spec);
