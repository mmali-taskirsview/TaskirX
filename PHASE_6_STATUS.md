# Phase 6: Advanced Features - Status

## 1. Advanced Targeting (Geo-fencing)
- [x] **Backend API**: Add `lat`, `lon`, and `radius` fields to Campaign targeting entity and DTOs.
- [x] **Frontend**: Update Campaign Wizard to allow selecting a location on a map (or inputting coordinates).
- [x] **Bidding Engine**: Implement Haversine formula in Go to filter bids based on user's device location vs campaign target.
- [x] **Verification**: Test with a bid request containing specific coordinates.

## 2. Header Bidding (Prebid.js)
- [x] **Adapter**: Create a `taskirxBidAdapter.js` that complies with Prebid.org spec.
- [x] **Endpoint**: Ensure `POST /bid` accepts payloads from Prebid adapter.
- [x] **Test Page**: Create a simple HTML page with Prebid.js to simulate a publisher site.

## 3. Video & Native Ads
- [x] **VAST Support**: Update Bidding Engine to return XML VAST responses.
- [x] **Native Assets**: Add fields for Icon, Image, Title, Description to Campaign creative model.
