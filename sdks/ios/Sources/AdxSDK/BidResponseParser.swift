import Foundation

internal class BidResponseParser {
    
    static func parse(bidResponse: BidResponse, adFormat: AdFormat) -> AdResponse? {
        guard let seatBid = bidResponse.seatbid?.first,
              let bid = seatBid.bid.first else {
            return nil
        }
        
        switch adFormat {
        case .banner, .interstitial:
            return parseBanner(bid: bid, bidId: bidResponse.bidid ?? bid.id)
        case .video, .rewardedVideo:
            return parseVideo(bid: bid, bidId: bidResponse.bidid ?? bid.id)
        case .native:
            return parseNative(bid: bid, bidId: bidResponse.bidid ?? bid.id)
        }
    }
    
    private static func parseBanner(bid: Bid, bidId: String) -> AdResponse {
        let markup = bid.adm
        let imageURL = extractImageURL(from: markup)
        let clickURL = extractClickURL(from: markup)
        
        return AdResponse(
            bidId: bidId,
            impressionId: bid.impid,
            price: bid.price,
            adFormat: .banner,
            imageURL: imageURL,
            clickURL: clickURL,
            impressionURL: bid.nurl.flatMap { URL(string: $0) },
            videoURL: nil,
            title: nil,
            description: nil,
            callToAction: nil,
            iconURL: nil,
            sponsoredBy: nil,
            width: bid.w,
            height: bid.h
        )
    }
    
    private static func parseVideo(bid: Bid, bidId: String) -> AdResponse {
        let markup = bid.adm
        let videoURL = extractVideoURL(from: markup)
        let clickURL = extractClickURL(from: markup)
        
        return AdResponse(
            bidId: bidId,
            impressionId: bid.impid,
            price: bid.price,
            adFormat: .video,
            imageURL: nil,
            clickURL: clickURL,
            impressionURL: bid.nurl.flatMap { URL(string: $0) },
            videoURL: videoURL,
            title: nil,
            description: nil,
            callToAction: nil,
            iconURL: nil,
            sponsoredBy: nil,
            width: bid.w,
            height: bid.h
        )
    }
    
    private static func parseNative(bid: Bid, bidId: String) -> AdResponse {
        guard let data = bid.adm.data(using: .utf8),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let native = json["native"] as? [String: Any] else {
            return AdResponse(
                bidId: bidId,
                impressionId: bid.impid,
                price: bid.price,
                adFormat: .native,
                imageURL: nil,
                clickURL: nil,
                impressionURL: bid.nurl.flatMap { URL(string: $0) },
                videoURL: nil,
                title: "Ad",
                description: "Sponsored content",
                callToAction: "Learn More",
                iconURL: nil,
                sponsoredBy: nil,
                width: bid.w,
                height: bid.h
            )
        }
        
        var title: String?
        var description: String?
        var imageURL: URL?
        var iconURL: URL?
        var clickURL: URL?
        var callToAction: String?
        
        // Parse assets
        if let assets = native["assets"] as? [[String: Any]] {
            for asset in assets {
                let id = asset["id"] as? Int ?? 0
                
                switch id {
                case 1: // Title
                    if let titleData = asset["title"] as? [String: Any],
                       let text = titleData["text"] as? String {
                        title = text
                    }
                case 2: // Main image
                    if let imgData = asset["img"] as? [String: Any],
                       let urlString = imgData["url"] as? String {
                        imageURL = URL(string: urlString)
                    }
                case 3: // Description
                    if let dataObj = asset["data"] as? [String: Any],
                       let value = dataObj["value"] as? String {
                        description = value
                    }
                default:
                    break
                }
            }
        }
        
        // Parse link
        if let link = native["link"] as? [String: Any],
           let urlString = link["url"] as? String {
            clickURL = URL(string: urlString)
        }
        
        return AdResponse(
            bidId: bidId,
            impressionId: bid.impid,
            price: bid.price,
            adFormat: .native,
            imageURL: imageURL,
            clickURL: clickURL,
            impressionURL: bid.nurl.flatMap { URL(string: $0) },
            videoURL: nil,
            title: title,
            description: description,
            callToAction: callToAction ?? "Learn More",
            iconURL: iconURL,
            sponsoredBy: nil,
            width: bid.w,
            height: bid.h
        )
    }
    
    // MARK: - Helper Methods
    
    private static func extractImageURL(from markup: String) -> URL? {
        // Extract image URL from HTML
        let pattern = #"<img[^>]+src=["']([^"']+)["']"#
        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: markup, range: NSRange(markup.startIndex..., in: markup)),
              match.numberOfRanges > 1,
              let range = Range(match.range(at: 1), in: markup) else {
            return URL(string: markup) // Fallback: treat as direct URL
        }
        
        return URL(string: String(markup[range]))
    }
    
    private static func extractClickURL(from markup: String) -> URL? {
        // Extract click URL from HTML
        let pattern = #"<a[^>]+href=["']([^"']+)["']"#
        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: markup, range: NSRange(markup.startIndex..., in: markup)),
              match.numberOfRanges > 1,
              let range = Range(match.range(at: 1), in: markup) else {
            return nil
        }
        
        return URL(string: String(markup[range]))
    }
    
    private static func extractVideoURL(from markup: String) -> URL? {
        // Extract video URL from VAST XML or direct URL
        let pattern = #"<MediaFile[^>]*>([^<]+)</MediaFile>"#
        guard let regex = try? NSRegularExpression(pattern: pattern),
              let match = regex.firstMatch(in: markup, range: NSRange(markup.startIndex..., in: markup)),
              match.numberOfRanges > 1,
              let range = Range(match.range(at: 1), in: markup) else {
            // Try as direct URL
            if markup.contains(".mp4") || markup.contains(".mov") {
                return URL(string: markup.trimmingCharacters(in: .whitespacesAndNewlines))
            }
            return nil
        }
        
        let urlString = String(markup[range]).trimmingCharacters(in: .whitespacesAndNewlines)
        return URL(string: urlString)
    }
}
