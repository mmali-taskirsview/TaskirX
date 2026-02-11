import Foundation

internal class UserIdManager {
    
    private let userDefaults = UserDefaults.standard
    private let userIdKey = "com.taskirx.sdk.userId"
    private let customUserIdKey = "com.taskirx.sdk.customUserId"
    
    // MARK: - User ID Management
    
    func getUserId() -> String {
        // Check for custom user ID first
        if let customId = userDefaults.string(forKey: customUserIdKey), !customId.isEmpty {
            return customId
        }
        
        // Get or generate SDK user ID
        if let userId = userDefaults.string(forKey: userIdKey), !userId.isEmpty {
            return userId
        }
        
        // Generate new user ID
        let newUserId = UUID().uuidString
        userDefaults.set(newUserId, forKey: userIdKey)
        return newUserId
    }
    
    func setCustomUserId(_ userId: String) {
        userDefaults.set(userId, forKey: customUserIdKey)
    }
    
    func clearCustomUserId() {
        userDefaults.removeObject(forKey: customUserIdKey)
    }
    
    func resetUserId() {
        userDefaults.removeObject(forKey: userIdKey)
        userDefaults.removeObject(forKey: customUserIdKey)
    }
}
