import Foundation

// MARK: - Custom Error

public enum TaskirXError: LocalizedError {
    case networkError(String)
    case decodingError(String)
    case httpError(statusCode: Int, message: String)
    case invalidResponse
    case timeout
    case retryExhausted
    
    public var errorDescription: String? {
        switch self {
        case .networkError(let message):
            return "Network Error: \(message)"
        case .decodingError(let message):
            return "Decoding Error: \(message)"
        case .httpError(let statusCode, let message):
            return "HTTP \(statusCode): \(message)"
        case .invalidResponse:
            return "Invalid response from server"
        case .timeout:
            return "Request timeout"
        case .retryExhausted:
            return "Max retries exhausted"
        }
    }
}

// MARK: - Request Manager

public class RequestManager {
    private let config: ClientConfig
    private let session: URLSession
    private var authToken: String?
    private let tokenLock = NSLock()
    
    public init(config: ClientConfig) {
        self.config = config
        
        let sessionConfig = URLSessionConfiguration.default
        sessionConfig.timeoutIntervalForRequest = config.timeout
        sessionConfig.timeoutIntervalForResource = config.timeout * 2
        self.session = URLSession(configuration: sessionConfig)
    }
    
    public func setAuthToken(_ token: String) {
        tokenLock.lock()
        defer { tokenLock.unlock() }
        self.authToken = token
    }
    
    public func clearAuthToken() {
        tokenLock.lock()
        defer { tokenLock.unlock() }
        self.authToken = nil
    }
    
    // MARK: - HTTP Methods
    
    public func get<T: Decodable>(_ endpoint: String) async throws -> T {
        var urlComponents = URLComponents(string: config.apiUrl + endpoint)!
        return try await request(urlComponents.url!, method: "GET")
    }
    
    public func post<T: Decodable>(_ endpoint: String, body: Encodable? = nil) async throws -> T {
        let url = URL(string: config.apiUrl + endpoint)!
        return try await request(url, method: "POST", body: body)
    }
    
    public func put<T: Decodable>(_ endpoint: String, body: Encodable? = nil) async throws -> T {
        let url = URL(string: config.apiUrl + endpoint)!
        return try await request(url, method: "PUT", body: body)
    }
    
    public func delete<T: Decodable>(_ endpoint: String) async throws -> T {
        let url = URL(string: config.apiUrl + endpoint)!
        return try await request(url, method: "DELETE")
    }
    
    // MARK: - Core Request Logic
    
    private func request<T: Decodable>(_ url: URL, method: String, body: Encodable? = nil) async throws -> T {
        var lastError: Error?
        let delays = [100, 300, 900] // milliseconds for exponential backoff
        
        for attempt in 0..<config.retryAttempts {
            do {
                var request = URLRequest(url: url)
                request.httpMethod = method
                request.timeoutInterval = config.timeout
                
                // Add headers
                request.setValue("application/json", forHTTPHeaderField: "Content-Type")
                request.setValue("application/json", forHTTPHeaderField: "Accept")
                request.setValue(config.apiKey, forHTTPHeaderField: "X-API-Key")
                request.setValue(UUID().uuidString, forHTTPHeaderField: "X-Request-ID")
                request.setValue("TaskirX-iOS/1.0", forHTTPHeaderField: "User-Agent")
                
                // Add auth token if available
                if let token = authToken {
                    request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
                }
                
                // Add body if provided
                if let body = body {
                    request.httpBody = try JSONEncoder().encode(body)
                }
                
                let (data, response) = try await session.data(for: request)
                
                // Check HTTP status
                guard let httpResponse = response as? HTTPURLResponse else {
                    throw TaskirXError.invalidResponse
                }
                
                guard (200..<300).contains(httpResponse.statusCode) else {
                    if let errorResponse = try? JSONDecoder().decode(ErrorResponse.self, from: data) {
                        throw TaskirXError.httpError(statusCode: httpResponse.statusCode, message: errorResponse.message)
                    }
                    throw TaskirXError.httpError(statusCode: httpResponse.statusCode, message: "HTTP Error")
                }
                
                // Decode response
                let decoder = JSONDecoder()
                decoder.keyDecodingStrategy = .convertFromSnakeCase
                let decodedResponse = try decoder.decode(ApiResponse<T>.self, from: data)
                
                if decodedResponse.success, let data = decodedResponse.data {
                    if config.debug {
                        print("[TaskirX] ✅ \(method) \(url.path) - Success")
                    }
                    return data
                } else if let error = decodedResponse.error {
                    throw TaskirXError.httpError(statusCode: httpResponse.statusCode, message: error.message)
                } else {
                    throw TaskirXError.decodingError("Invalid response structure")
                }
                
            } catch let error as TaskirXError {
                lastError = error
                
                if attempt < config.retryAttempts - 1 {
                    let delayMs = delays[min(attempt, delays.count - 1)]
                    if config.debug {
                        print("[TaskirX] ⚠️ Retry attempt \(attempt + 1) after \(delayMs)ms")
                    }
                    try await Task.sleep(nanoseconds: UInt64(delayMs * 1_000_000))
                } else {
                    if config.debug {
                        print("[TaskirX] ❌ \(method) \(url.path) - \(error)")
                    }
                }
            } catch {
                lastError = TaskirXError.networkError(error.localizedDescription)
                
                if attempt < config.retryAttempts - 1 {
                    let delayMs = delays[min(attempt, delays.count - 1)]
                    if config.debug {
                        print("[TaskirX] ⚠️ Retry attempt \(attempt + 1) after \(delayMs)ms")
                    }
                    try await Task.sleep(nanoseconds: UInt64(delayMs * 1_000_000))
                }
            }
        }
        
        throw lastError ?? TaskirXError.retryExhausted
    }
}
