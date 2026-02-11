import Foundation

/// OpenRTB 2.5 Data Models

// MARK: - Bid Request

internal struct BidRequest: Codable {
    let id: String
    let imp: [Impression]
    let device: Device
    let user: User
    let app: App
    let test: Int
}

internal struct Impression: Codable {
    let id: String
    let banner: Banner?
    let video: Video?
    let native: Native?
    let bidfloor: Double
    let tagid: String
}

internal struct Banner: Codable {
    let w: Int
    let h: Int
    let format: [Format]?
}

internal struct Format: Codable {
    let w: Int
    let h: Int
}

internal struct Video: Codable {
    let mimes: [String]
    let minduration: Int
    let maxduration: Int
    let protocols: [Int]
    let w: Int
    let h: Int
}

internal struct Native: Codable {
    let request: String
}

internal struct Device: Codable {
    let ua: String
    let ip: String
    let devicetype: Int
    let make: String
    let model: String
    let os: String
    let osv: String
    let w: Int
    let h: Int
    let ifa: String?
    let lmt: Int
    let connectiontype: Int
    let language: String
}

internal struct User: Codable {
    let id: String
    let yob: Int?
    let gender: String?
    let keywords: String?
    let geo: Geo?
}

internal struct Geo: Codable {
    let lat: Double
    let lon: Double
    let type: Int
    let accuracy: Int?
    let country: String?
    let city: String?
}

internal struct App: Codable {
    let id: String
    let name: String
    let bundle: String
    let storeurl: String?
    let publisher: Publisher
}

internal struct Publisher: Codable {
    let id: String
    let name: String?
}

// MARK: - Bid Response

internal struct BidResponse: Codable {
    let id: String
    let seatbid: [SeatBid]?
    let bidid: String?
    let cur: String
    let nbr: Int?
}

internal struct SeatBid: Codable {
    let bid: [Bid]
    let seat: String?
}

internal struct Bid: Codable {
    let id: String
    let impid: String
    let price: Double
    let adid: String?
    let nurl: String?
    let adm: String
    let adomain: [String]?
    let crid: String?
    let w: Int?
    let h: Int?
    let ext: [String: AnyCodable]?
}

// MARK: - Ad Response (Parsed for SDK use)

internal struct AdResponse {
    let bidId: String
    let impressionId: String
    let price: Double
    let adFormat: AdFormat
    let imageURL: URL?
    let clickURL: URL?
    let impressionURL: URL?
    let videoURL: URL?
    let title: String?
    let description: String?
    let callToAction: String?
    let iconURL: URL?
    let sponsoredBy: String?
    let width: Int?
    let height: Int?
}

// MARK: - Helper for dynamic JSON

internal struct AnyCodable: Codable {
    let value: Any
    
    init(_ value: Any) {
        self.value = value
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if let string = try? container.decode(String.self) {
            value = string
        } else if let int = try? container.decode(Int.self) {
            value = int
        } else if let double = try? container.decode(Double.self) {
            value = double
        } else if let bool = try? container.decode(Bool.self) {
            value = bool
        } else if let array = try? container.decode([AnyCodable].self) {
            value = array.map { $0.value }
        } else if let dict = try? container.decode([String: AnyCodable].self) {
            value = dict.mapValues { $0.value }
        } else {
            value = NSNull()
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        switch value {
        case let string as String:
            try container.encode(string)
        case let int as Int:
            try container.encode(int)
        case let double as Double:
            try container.encode(double)
        case let bool as Bool:
            try container.encode(bool)
        case let array as [Any]:
            try container.encode(array.map { AnyCodable($0) })
        case let dict as [String: Any]:
            try container.encode(dict.mapValues { AnyCodable($0) })
        default:
            try container.encodeNil()
        }
    }
}
