import Foundation

internal class NetworkClient {
    
    private var configuration: AdxConfiguration!
    private var session: URLSession!
    
    func configure(with configuration: AdxConfiguration) {
        self.configuration = configuration
        
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = configuration.connectionTimeout
        config.timeoutIntervalForResource = configuration.readTimeout
        
        self.session = URLSession(configuration: config)
    }
    
    // MARK: - Bid Request
    
    func requestAd(bidRequest: BidRequest) async throws -> BidResponse {
        guard let url = URL(string: "\(configuration.apiEndpoint)/api/rtb/bid-request") else {
            throw AdxError.invalidRequest("Invalid API endpoint")
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(AdxSDK.version, forHTTPHeaderField: "X-ADX-SDK-Version")
        request.setValue(AdxSDK.shared.publisherId, forHTTPHeaderField: "X-ADX-Publisher-Id")
        request.setValue("iOS", forHTTPHeaderField: "X-taskirx")
        
        let encoder = JSONEncoder()
        request.httpBody = try encoder.encode(bidRequest)
        
        if configuration.enableDebug {
            if let jsonString = String(data: request.httpBody!, encoding: .utf8) {
                AdxSDK.shared.log("Bid Request: \(jsonString)")
            }
        }
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw AdxError.networkError(NSError(domain: "Invalid response", code: -1))
        }
        
        guard httpResponse.statusCode == 200 else {
            throw AdxError.networkError(NSError(domain: "HTTP \(httpResponse.statusCode)", code: httpResponse.statusCode))
        }
        
        if configuration.enableDebug {
            if let jsonString = String(data: data, encoding: .utf8) {
                AdxSDK.shared.log("Bid Response: \(jsonString)")
            }
        }
        
        let decoder = JSONDecoder()
        let bidResponse = try decoder.decode(BidResponse.self, from: data)
        
        return bidResponse
    }
    
    // MARK: - Tracking
    
    func trackImpression(bidId: String) async throws {
        guard let url = URL(string: "\(configuration.apiEndpoint)/api/rtb/impression/\(bidId)") else {
            throw AdxError.invalidRequest("Invalid impression URL")
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        let (_, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AdxError.networkError(NSError(domain: "Impression tracking failed", code: -1))
        }
        
        AdxSDK.shared.log("Impression tracked: \(bidId)")
    }
    
    func trackClick(bidId: String) async throws {
        guard let url = URL(string: "\(configuration.apiEndpoint)/api/rtb/click/\(bidId)") else {
            throw AdxError.invalidRequest("Invalid click URL")
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        let (_, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AdxError.networkError(NSError(domain: "Click tracking failed", code: -1))
        }
        
        AdxSDK.shared.log("Click tracked: \(bidId)")
    }
    
    func trackConversion(bidId: String, eventType: String, value: Double? = nil) async throws {
        var urlString = "\(configuration.apiEndpoint)/api/rtb/conversion/\(bidId)?eventType=\(eventType)"
        if let value = value {
            urlString += "&value=\(value)"
        }
        
        guard let url = URL(string: urlString) else {
            throw AdxError.invalidRequest("Invalid conversion URL")
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        let (_, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw AdxError.networkError(NSError(domain: "Conversion tracking failed", code: -1))
        }
        
        AdxSDK.shared.log("Conversion tracked: \(bidId) (\(eventType))")
    }
}
