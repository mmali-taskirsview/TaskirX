import SwiftUI

struct ContentView: View {
    
    @StateObject private var interstitial = AdxInterstitial(placementId: "placement-123")
    @StateObject private var rewardedVideo = AdxRewardedVideo(placementId: "placement-456")
    
    @State private var rewardMessage = ""
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 20) {
                    
                    // Banner Ad Example
                    bannerSection
                    
                    Divider()
                    
                    // Native Ad Example
                    nativeSection
                    
                    Divider()
                    
                    // Interstitial Ad Example
                    interstitialSection
                    
                    Divider()
                    
                    // Rewarded Video Example
                    rewardedVideoSection
                    
                    if !rewardMessage.isEmpty {
                        Text(rewardMessage)
                            .font(.headline)
                            .foregroundColor(.green)
                            .padding()
                    }
                }
                .padding()
            }
            .navigationTitle("AdxSDK Examples")
        }
        .adxInterstitial(interstitial)
        .adxRewardedVideo(rewardedVideo)
    }
    
    // MARK: - Banner Section
    
    private var bannerSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Banner Ad")
                .font(.title2)
                .fontWeight(.bold)
            
            AdxBannerView(
                placementId: "banner-placement-123",
                adSize: .banner320x50,
                onAdLoaded: {
                    print("Banner ad loaded")
                },
                onAdFailed: { error in
                    print("Banner ad failed: \(error)")
                },
                onAdClicked: {
                    print("Banner ad clicked")
                }
            )
        }
    }
    
    // MARK: - Native Section
    
    private var nativeSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Native Ad")
                .font(.title2)
                .fontWeight(.bold)
            
            AdxNativeView(
                placementId: "native-placement-456",
                onAdLoaded: {
                    print("Native ad loaded")
                },
                onAdFailed: { error in
                    print("Native ad failed: \(error)")
                },
                onAdClicked: {
                    print("Native ad clicked")
                }
            )
        }
    }
    
    // MARK: - Interstitial Section
    
    private var interstitialSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Interstitial Ad")
                .font(.title2)
                .fontWeight(.bold)
            
            HStack(spacing: 15) {
                Button("Load Interstitial") {
                    interstitial.loadAd()
                }
                .buttonStyle(.bordered)
                
                Button("Show Interstitial") {
                    interstitial.show()
                }
                .buttonStyle(.borderedProminent)
                .disabled(!interstitial.isReady)
            }
            
            Text(interstitial.isReady ? "✓ Ready to show" : "Loading...")
                .font(.caption)
                .foregroundColor(interstitial.isReady ? .green : .secondary)
        }
        .onAppear {
            setupInterstitialCallbacks()
        }
    }
    
    // MARK: - Rewarded Video Section
    
    private var rewardedVideoSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Rewarded Video")
                .font(.title2)
                .fontWeight(.bold)
            
            HStack(spacing: 15) {
                Button("Load Rewarded Video") {
                    rewardedVideo.loadAd()
                }
                .buttonStyle(.bordered)
                
                Button("Show Rewarded Video") {
                    rewardedVideo.show()
                }
                .buttonStyle(.borderedProminent)
                .disabled(!rewardedVideo.isReady)
            }
            
            Text(rewardedVideo.isReady ? "✓ Ready to show" : "Loading...")
                .font(.caption)
                .foregroundColor(rewardedVideo.isReady ? .green : .secondary)
        }
        .onAppear {
            setupRewardedVideoCallbacks()
        }
    }
    
    // MARK: - Setup Callbacks
    
    private func setupInterstitialCallbacks() {
        interstitial.onAdLoaded = {
            print("✅ Interstitial ad loaded")
        }
        
        interstitial.onAdFailed = { error in
            print("❌ Interstitial ad failed: \(error)")
        }
        
        interstitial.onAdDismissed = {
            print("Interstitial ad dismissed")
        }
    }
    
    private func setupRewardedVideoCallbacks() {
        rewardedVideo.onAdLoaded = {
            print("✅ Rewarded video loaded")
        }
        
        rewardedVideo.onAdFailed = { error in
            print("❌ Rewarded video failed: \(error)")
        }
        
        rewardedVideo.onRewardEarned = { reward in
            rewardMessage = "🎉 Earned \(reward.amount) \(reward.type)!"
            print("🎉 Reward earned: \(reward.type) x\(reward.amount)")
            
            // Clear message after 3 seconds
            DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                rewardMessage = ""
            }
        }
        
        rewardedVideo.onAdDismissed = {
            print("Rewarded video dismissed")
        }
    }
}

#Preview {
    ContentView()
}
