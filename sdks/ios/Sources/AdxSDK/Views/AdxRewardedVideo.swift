import SwiftUI
import AVKit

/// SwiftUI rewarded video ad manager
@MainActor
public class AdxRewardedVideo: ObservableObject {
    
    @Published public private(set) var isReady = false
    @Published public private(set) var isPresented = false
    
    private let placementId: String
    private let targeting: AdxTargeting?
    private var currentAd: AdResponse?
    private var didEarnReward = false
    
    public var onAdLoaded: (() -> Void)?
    public var onAdFailed: ((Error) -> Void)?
    public var onRewardEarned: ((AdxReward) -> Void)?
    public var onAdDismissed: (() -> Void)?
    
    public init(placementId: String, targeting: AdxTargeting? = nil) {
        self.placementId = placementId
        self.targeting = targeting
    }
    
    public func loadAd() {
        Task {
            do {
                try ensureInitialized()
                
                let sdk = AdxSDK.shared
                let networkClient = sdk.getNetworkClient()
                let deviceInfo = sdk.getDeviceInfo()
                
                // Build bid request
                let bidRequestBuilder = BidRequestBuilder(deviceInfo: deviceInfo)
                let bidRequest = bidRequestBuilder.buildRequest(
                    placementId: placementId,
                    adFormat: .rewardedVideo,
                    adSize: nil,
                    targeting: targeting
                )
                
                // Request ad
                let bidResponse = try await networkClient.requestAd(bidRequest: bidRequest)
                
                // Parse response
                if let ad = BidResponseParser.parse(bidResponse: bidResponse, adFormat: .rewardedVideo) {
                    currentAd = ad
                    isReady = true
                    didEarnReward = false
                    onAdLoaded?()
                    sdk.log("Rewarded video loaded")
                } else {
                    throw AdxError.noFill
                }
            } catch {
                onAdFailed?(error)
                AdxSDK.shared.log("Rewarded video load failed: \(error)", level: .error)
            }
        }
    }
    
    public func show() {
        guard isReady, let ad = currentAd else {
            AdxSDK.shared.log("Rewarded video not ready", level: .warning)
            return
        }
        
        // Track impression
        Task {
            try? await AdxSDK.shared.getNetworkClient().trackImpression(bidId: ad.bidId)
        }
        
        isPresented = true
    }
    
    internal func handleVideoComplete() {
        guard !didEarnReward else { return }
        
        didEarnReward = true
        let reward = AdxReward(type: "coins", amount: 1)
        onRewardEarned?(reward)
        
        // Track conversion
        if let ad = currentAd {
            Task {
                try? await AdxSDK.shared.getNetworkClient().trackConversion(bidId: ad.bidId)
            }
        }
    }
    
    internal func handleClick() {
        guard let ad = currentAd else { return }
        
        Task {
            try? await AdxSDK.shared.getNetworkClient().trackClick(bidId: ad.bidId)
            
            if let clickURL = ad.clickURL {
                await MainActor.run {
                    #if canImport(UIKit)
                    UIApplication.shared.open(clickURL)
                    #endif
                }
            }
        }
    }
    
    internal func dismiss() {
        isPresented = false
        isReady = false
        currentAd = nil
        onAdDismissed?()
    }
}

// MARK: - Rewarded Video View

struct RewardedVideoView: View {
    
    @ObservedObject var rewardedVideo: AdxRewardedVideo
    let ad: AdResponse
    @State private var player: AVPlayer?
    @State private var videoCompleted = false
    @State private var canClose = false
    
    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()
            
            if let videoURL = ad.videoURL {
                VideoPlayer(player: player)
                    .ignoresSafeArea()
                    .onAppear {
                        setupPlayer(videoURL: videoURL)
                    }
                    .onDisappear {
                        player?.pause()
                    }
            } else {
                VStack {
                    Text("Video ad")
                        .font(.title)
                        .foregroundColor(.white)
                }
            }
            
            VStack {
                HStack {
                    Spacer()
                    if canClose {
                        Button(action: { rewardedVideo.dismiss() }) {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundColor(.white)
                                .font(.title)
                                .padding()
                        }
                    } else {
                        Text("Video will end soon...")
                            .font(.caption)
                            .foregroundColor(.white)
                            .padding()
                            .background(Color.black.opacity(0.5))
                            .cornerRadius(8)
                            .padding()
                    }
                }
                Spacer()
            }
        }
    }
    
    private func setupPlayer(videoURL: URL) {
        player = AVPlayer(url: videoURL)
        
        // Observe when video finishes
        NotificationCenter.default.addObserver(
            forName: .AVPlayerItemDidPlayToEndTime,
            object: player?.currentItem,
            queue: .main
        ) { _ in
            videoCompleted = true
            canClose = true
            rewardedVideo.handleVideoComplete()
        }
        
        // Allow close after 5 seconds
        DispatchQueue.main.asyncAfter(deadline: .now() + 5) {
            canClose = true
        }
        
        player?.play()
    }
}

// MARK: - View Modifier

public extension View {
    func adxRewardedVideo(_ rewardedVideo: AdxRewardedVideo) -> some View {
        self.fullScreenCover(isPresented: Binding(
            get: { rewardedVideo.isPresented },
            set: { if !$0 { rewardedVideo.dismiss() } }
        )) {
            if let ad = rewardedVideo.currentAd {
                RewardedVideoView(rewardedVideo: rewardedVideo, ad: ad)
            }
        }
    }
}
