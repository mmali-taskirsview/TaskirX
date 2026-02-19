// import { registerBidder } from './prebid/src/adapters/bidderFactory.js';
// import { BANNER, NATIVE, VIDEO } from './prebid/src/mediaTypes.js';

// Prebid Constants for standalone usage
const BANNER = 'banner';
const NATIVE = 'native';
const VIDEO = 'video';

const BIDDER_CODE = 'taskirx';
const ENDPOINT_URL = 'http://localhost:8080/bid'; // Local Dev Endpoint
// const ENDPOINT_URL = 'https://bidding.taskirx.com/bid'; // Production Endpoint

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
    return !!(bid.params && bid.params.placementId);
  },

  /**
   * Make a server request from the list of BidRequests.
   *
   * @param {Array} validBidRequests A non-empty list of valid bid requests.
   * @param {object} bidderRequest
   * @return ServerRequest Info describing the request to the server.
   */
  buildRequests: function(validBidRequests, bidderRequest) {
    return validBidRequests.map(bid => {
      // Safely extract dimensions for banner
      let dimensions = [0, 0];
      if (bid.mediaTypes && bid.mediaTypes.banner && bid.mediaTypes.banner.sizes) {
          dimensions = bid.mediaTypes.banner.sizes[0];
      }

      // Build PMP Object if Deals are present
      let pmp = null;
      if (bid.params && bid.params.deals && Array.isArray(bid.params.deals)) {
        pmp = {
            private_auction: bid.params.privateAuction ? 1 : 0,
            deals: bid.params.deals.map(d => ({
                id: d.id,
                bid_floor: d.floor || 0
            }))
        };
      }

      const payload = {
        id: bid.bidId, // Use bidId for tracking individual bids
        timestamp: new Date().toISOString(),
        publisher_id: bid.params.publisherId || 'unknown-publisher',
        ad_slot: {
          id: bid.adUnitCode,
          dimensions: dimensions,
          position: bid.params.position || "unknown", 
          formats: Object.keys(bid.mediaTypes) // ["banner", "native", "video"]
        },
        pmp: pmp, // Add PMP object (Private Marketplace)
        user: {
          // Sync user Ids if available
          id: 'user-' + Math.random().toString(36).substring(7),
          // gdpr_consent: bidderRequest.gdprConsent ? ...
        },
        device: {
          type: "desktop", // simplified
          ua: navigator.userAgent,
          geo: { lat: 0, lon: 0 } // placeholder
        },
        context: {
          referer: bidderRequest.refererInfo ? bidderRequest.refererInfo.page : window.location.href,
          page_url: bidderRequest.refererInfo ? bidderRequest.refererInfo.page : window.location.href
        }
      };

      return {
        method: 'POST',
        url: ENDPOINT_URL,
        data: JSON.stringify(payload),
        bidId: bid.bidId // pass through
      };
    });
  },

  /**
   * Unpack the response from the server into a list of bids.
   *
   * @param {object} serverResponse A successful response from the server.
   * @param {object} originalRequest The original request object.
   * @return {Array} An array of bids which were nested inside the server.
   */
  interpretResponse: function(serverResponse, originalRequest) {
    const serverBody = serverResponse.body;
    const bidResponses = [];

    if (!serverBody || !serverBody.bid_price || serverBody.bid_price <= 0) {
      return []; // no bid
    }

    const bid = {
      requestId: originalRequest.bidId, 
      cpm: serverBody.bid_price,
      currency: 'USD',
      dealId: serverBody.deal_id, // Pass Deal ID if returned
      netRevenue: true,
      ttl: serverBody.ttl || 300,
      creativeId: serverBody.creative_url || serverBody.request_id,
      meta: {
        advertiserDomains: [] 
      }
    };

    // Construct the ad display based on 'ad_markup' if present, otherwise build a simple tag
    if (serverBody.ad_markup) {
        bid.ad = serverBody.ad_markup;
    } else {
        // Fallback for banner if no markup provided
        bid.width = 300; // should come from response or request
        bid.height = 250;
        bid.ad = `<a href="${serverBody.click_url}" target="_blank">
                    <img src="${serverBody.creative_url}" width="${bid.width}" height="${bid.height}" />
                  </a>
                  <img src="${serverBody.impression_url}" style="display:none" />`;
    }
    
    // Handle Video Context
    if (originalRequest.data && JSON.parse(originalRequest.data).ad_slot.formats.includes('video')) {
        bid.vastXml = serverBody.ad_markup; // VAST XML for video
        bid.mediaType = VIDEO;
    }

    bidResponses.push(bid);
    return bidResponses;
  }
};

if (typeof registerBidder === 'function') {
  registerBidder(spec);
}
