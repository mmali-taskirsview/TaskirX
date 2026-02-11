import SwiftUI
import TaskirX

struct TaskirXExampleView: View {
    @State private var status: String = "Ready"
    @State private var authToken: String = ""
    let client = TaskirXClient.create(apiUrl: "http://localhost:3000", apiKey: "test-key", debug: true)
    
    var body: some View {
        VStack(spacing: 20) {
            Text("TaskirX SDK Examples")
                .font(.title)
                .bold()
            
            ScrollView {
                VStack(spacing: 15) {
                    // Health Check
                    Button("1. Health Check") {
                        Task {
                            status = "Checking health..."
                            let result = await client.getHealth()
                            result.onSuccess { _ in
                                status = "✅ Health check passed"
                            }
                            result.onFailure { error in
                                status = "❌ Health check failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Status
                    Button("2. Get Status") {
                        Task {
                            status = "Fetching status..."
                            let result = await client.getStatus()
                            result.onSuccess { _ in
                                status = "✅ Status retrieved"
                            }
                            result.onFailure { error in
                                status = "❌ Status failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Authentication
                    Button("3. Login") {
                        Task {
                            status = "Logging in..."
                            let loginResult = try? await client.auth.login(
                                email: "user@example.com",
                                password: "password123"
                            )
                            if let login = loginResult {
                                authToken = login.token
                                status = "✅ Logged in as \(login.user.name)"
                            } else {
                                status = "❌ Login failed"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Get Profile
                    Button("4. Get Profile") {
                        Task {
                            status = "Fetching profile..."
                            let result = await client.getProfile()
                            result.onSuccess { user in
                                status = "✅ Profile: \(user.name) (\(user.email))"
                            }
                            result.onFailure { error in
                                status = "❌ Profile failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Get Campaigns
                    Button("5. Get Campaigns") {
                        Task {
                            status = "Fetching campaigns..."
                            let result = await client.getCampaigns(limit: 10)
                            result.onSuccess { campaigns in
                                status = "✅ Found \(campaigns.count) campaigns"
                            }
                            result.onFailure { error in
                                status = "❌ Campaigns failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Create Campaign
                    Button("6. Create Campaign") {
                        Task {
                            status = "Creating campaign..."
                            let result = await client.createCampaign(
                                name: "iOS Test Campaign",
                                budget: 1000.0,
                                startDate: "2024-01-01",
                                endDate: "2024-01-31",
                                targetAudience: ["age": .string("18-35")]
                            )
                            result.onSuccess { campaign in
                                status = "✅ Campaign created: \(campaign.name)"
                            }
                            result.onFailure { error in
                                status = "❌ Campaign creation failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Get Analytics
                    Button("7. Get Analytics") {
                        Task {
                            status = "Fetching analytics..."
                            let result = await client.getRealtimeAnalytics()
                            result.onSuccess { analytics in
                                status = "✅ Impressions: \(analytics.impressions), Clicks: \(analytics.clicks)"
                            }
                            result.onFailure { error in
                                status = "❌ Analytics failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Submit Bid
                    Button("8. Submit Bid") {
                        Task {
                            status = "Submitting bid..."
                            let result = await client.submitBid(
                                campaignId: "camp-1",
                                adSlotId: "slot-1",
                                amount: 5.0
                            )
                            result.onSuccess { bid in
                                status = "✅ Bid submitted: $\(bid.amount)"
                            }
                            result.onFailure { error in
                                status = "❌ Bid failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Create Ad
                    Button("9. Create Ad") {
                        Task {
                            status = "Creating ad..."
                            let result = await client.createAd(
                                campaignId: "camp-1",
                                placement: "banner",
                                imageUrl: "https://example.com/ad.jpg",
                                clickUrl: "https://example.com/landing",
                                dimensions: "300x250"
                            )
                            result.onSuccess { ad in
                                status = "✅ Ad created: \(ad.id)"
                            }
                            result.onFailure { error in
                                status = "❌ Ad creation failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Subscribe Webhook
                    Button("10. Subscribe Webhook") {
                        Task {
                            status = "Subscribing to webhook..."
                            let result = await client.subscribeWebhook(
                                url: "https://example.com/webhook",
                                events: ["campaign.created", "bid.submitted"]
                            )
                            result.onSuccess { webhook in
                                status = "✅ Webhook subscribed: \(webhook.id)"
                            }
                            result.onFailure { error in
                                status = "❌ Webhook subscription failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Dashboard
                    Button("11. Get Dashboard") {
                        Task {
                            status = "Fetching dashboard..."
                            let result = await client.getDashboard()
                            result.onSuccess { _ in
                                status = "✅ Dashboard loaded"
                            }
                            result.onFailure { error in
                                status = "❌ Dashboard failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Get Statistics
                    Button("12. Get Statistics") {
                        Task {
                            status = "Fetching statistics..."
                            let result = await client.getStatistics()
                            result.onSuccess { stats in
                                status = "✅ Statistics loaded"
                            }
                            result.onFailure { error in
                                status = "❌ Statistics failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Logout
                    Button("13. Logout") {
                        Task {
                            status = "Logging out..."
                            let result = await client.logout()
                            result.onSuccess { _ in
                                authToken = ""
                                status = "✅ Logged out"
                            }
                            result.onFailure { error in
                                status = "❌ Logout failed: \(error)"
                            }
                        }
                    }
                    .buttonStyle(.bordered)
                    
                    // Enable Debug
                    Button("14. Enable Debug") {
                        client.enableDebug(true)
                        status = "✅ Debug mode enabled"
                    }
                    .buttonStyle(.bordered)
                }
                .padding()
            }
            
            // Status Display
            VStack(alignment: .leading) {
                Text("Status:")
                    .font(.caption)
                    .foregroundColor(.gray)
                Text(status)
                    .font(.body)
                    .lineLimit(3)
                    .padding(8)
                    .background(Color.gray.opacity(0.2))
                    .cornerRadius(6)
            }
            .padding()
        }
        .padding()
    }
}

#Preview {
    TaskirXExampleView()
}
