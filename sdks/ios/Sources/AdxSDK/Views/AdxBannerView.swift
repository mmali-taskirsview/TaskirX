import SwiftUI

/// SwiftUI banner ad view
public struct AdxBannerView: View {
    
    private let placementId: String
    private let adSize: AdSize
    private let targeting: AdxTargeting?
    private let onAdLoaded: (() -> Void)?
    private let onAdFailed: ((Error) -> Void)?
    private let onAdClicked: (() -> Void)?
    
    @StateObject private var viewModel: BannerViewModel
    
    public init(
        placementId: String,
        adSize: AdSize,
        targeting: AdxTargeting? = nil,
        onAdLoaded: (() -> Void)? = nil,
        onAdFailed: ((Error) -> Void)? = nil,
        onAdClicked: (() -> Void)? = nil
    ) {
        self.placementId = placementId
        self.adSize = adSize
        self.targeting = targeting
        self.onAdLoaded = onAdLoaded
        self.onAdFailed = onAdFailed
        self.onAdClicked = onAdClicked
        
        _viewModel = StateObject(wrappedValue: BannerViewModel(
            placementId: placementId,
            adSize: adSize,
            targeting: targeting
        ))
    }
    
    public var body: some View {
        Group {
            if viewModel.isLoading {
                ProgressView()
                    .frame(width: adSize.size.width, height: adSize.size.height)
            } else if let ad = viewModel.currentAd {
                bannerContent(ad: ad)
            } else {
                Color.clear
                    .frame(width: adSize.size.width, height: adSize.size.height)
            }
        }
        .frame(width: adSize.size.width, height: adSize.size.height)
        .onAppear {
            viewModel.loadAd()
        }
        .onChange(of: viewModel.currentAd) { newAd in
            if newAd != nil {
                onAdLoaded?()
            }
        }
        .onChange(of: viewModel.error) { error in
            if let error = error {
                onAdFailed?(error)
            }
        }
    }
    
    @ViewBuilder
    private func bannerContent(ad: AdResponse) -> some View {
        if let imageURL = ad.imageURL {
            AsyncImage(url: imageURL) { image in
                image
                    .resizable()
                    .aspectRatio(contentMode: .fit)
            } placeholder: {
                ProgressView()
            }
            .frame(width: adSize.size.width, height: adSize.size.height)
            .contentShape(Rectangle())
            .onTapGesture {
                viewModel.handleClick()
                onAdClicked?()
            }
        } else {
            Text("Ad")
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(Color.gray.opacity(0.2))
        }
    }
}

// MARK: - Banner ViewModel

@MainActor
class BannerViewModel: ObservableObject {
    @Published var currentAd: AdResponse?
    @Published var isLoading = false
    @Published var error: Error?
    
    private let placementId: String
    private let adSize: AdSize
    private let targeting: AdxTargeting?
    private var refreshTimer: Timer?
    
    init(placementId: String, adSize: AdSize, targeting: AdxTargeting?) {
        self.placementId = placementId
        self.adSize = adSize
        self.targeting = targeting
    }
    
    func loadAd() {
        guard !isLoading else { return }
        
        Task {
            isLoading = true
            error = nil
            
            do {
                try ensureInitialized()
                
                let sdk = AdxSDK.shared
                let networkClient = sdk.getNetworkClient()
                let deviceInfo = sdk.getDeviceInfo()
                
                // Build bid request
                let bidRequestBuilder = BidRequestBuilder(deviceInfo: deviceInfo)
                let bidRequest = bidRequestBuilder.buildRequest(
                    placementId: placementId,
                    adFormat: .banner,
                    adSize: adSize,
                    targeting: targeting
                )
                
                // Request ad
                let bidResponse = try await networkClient.requestAd(bidRequest: bidRequest)
                
                // Parse response
                if let ad = BidResponseParser.parse(bidResponse: bidResponse, adFormat: .banner) {
                    currentAd = ad
                    
                    // Track impression
                    if let impressionURL = ad.impressionURL {
                        try? await networkClient.trackImpression(bidId: ad.bidId)
                    }
                    
                    // Setup auto-refresh
                    setupAutoRefresh()
                } else {
                    throw AdxError.noFill
                }
                
                isLoading = false
            } catch {
                self.error = error
                isLoading = false
                AdxSDK.shared.log("Banner ad load failed: \(error)", level: .error)
            }
        }
    }
    
    func handleClick() {
        guard let ad = currentAd else { return }
        
        Task {
            // Track click
            try? await AdxSDK.shared.getNetworkClient().trackClick(bidId: ad.bidId)
            
            // Open click URL
            if let clickURL = ad.clickURL {
                await MainActor.run {
                    #if canImport(UIKit)
                    UIApplication.shared.open(clickURL)
                    #endif
                }
            }
        }
    }
    
    private func setupAutoRefresh() {
        let refreshInterval = AdxSDK.shared.getConfiguration().refreshInterval
        guard refreshInterval > 0 else { return }
        
        refreshTimer?.invalidate()
        refreshTimer = Timer.scheduledTimer(withTimeInterval: refreshInterval, repeats: true) { [weak self] _ in
            self?.loadAd()
        }
    }
    
    deinit {
        refreshTimer?.invalidate()
    }
}
