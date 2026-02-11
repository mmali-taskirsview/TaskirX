import SwiftUI

/// SwiftUI interstitial ad manager
@MainActor
public class AdxInterstitial: ObservableObject {
    
    @Published public private(set) var isReady = false
    @Published public private(set) var isPresented = false
    
    private let placementId: String
    private let targeting: AdxTargeting?
    private var currentAd: AdResponse?
    
    public var onAdLoaded: (() -> Void)?
    public var onAdFailed: ((Error) -> Void)?
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
                    adFormat: .interstitial,
                    adSize: nil,
                    targeting: targeting
                )
                
                // Request ad
                let bidResponse = try await networkClient.requestAd(bidRequest: bidRequest)
                
                // Parse response
                if let ad = BidResponseParser.parse(bidResponse: bidResponse, adFormat: .interstitial) {
                    currentAd = ad
                    isReady = true
                    onAdLoaded?()
                    sdk.log("Interstitial ad loaded")
                } else {
                    throw AdxError.noFill
                }
            } catch {
                onAdFailed?(error)
                AdxSDK.shared.log("Interstitial ad load failed: \(error)", level: .error)
            }
        }
    }
    
    public func show() {
        guard isReady, let ad = currentAd else {
            AdxSDK.shared.log("Interstitial ad not ready", level: .warning)
            return
        }
        
        // Track impression
        Task {
            try? await AdxSDK.shared.getNetworkClient().trackImpression(bidId: ad.bidId)
        }
        
        isPresented = true
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

// MARK: - Interstitial View

struct InterstitialAdView: View {
    
    @ObservedObject var interstitial: AdxInterstitial
    let ad: AdResponse
    
    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()
            
            VStack {
                HStack {
                    Spacer()
                    Button(action: { interstitial.dismiss() }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(.white)
                            .font(.title)
                            .padding()
                    }
                }
                
                Spacer()
                
                if let imageURL = ad.imageURL {
                    AsyncImage(url: imageURL) { image in
                        image
                            .resizable()
                            .aspectRatio(contentMode: .fit)
                    } placeholder: {
                        ProgressView()
                    }
                    .contentShape(Rectangle())
                    .onTapGesture {
                        interstitial.handleClick()
                    }
                } else {
                    Text("Advertisement")
                        .font(.largeTitle)
                        .foregroundColor(.white)
                }
                
                Spacer()
            }
        }
    }
}

// MARK: - View Modifier

public extension View {
    func adxInterstitial(_ interstitial: AdxInterstitial) -> some View {
        self.fullScreenCover(isPresented: Binding(
            get: { interstitial.isPresented },
            set: { if !$0 { interstitial.dismiss() } }
        )) {
            if let ad = interstitial.currentAd {
                InterstitialAdView(interstitial: interstitial, ad: ad)
            }
        }
    }
}
