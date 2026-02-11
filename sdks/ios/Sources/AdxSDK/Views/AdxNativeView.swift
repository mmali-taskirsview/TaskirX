import SwiftUI

/// SwiftUI native ad view
public struct AdxNativeView: View {
    
    private let placementId: String
    private let targeting: AdxTargeting?
    private let onAdLoaded: (() -> Void)?
    private let onAdFailed: ((Error) -> Void)?
    private let onAdClicked: (() -> Void)?
    
    @StateObject private var viewModel: NativeViewModel
    
    public init(
        placementId: String,
        targeting: AdxTargeting? = nil,
        onAdLoaded: (() -> Void)? = nil,
        onAdFailed: ((Error) -> Void)? = nil,
        onAdClicked: (() -> Void)? = nil
    ) {
        self.placementId = placementId
        self.targeting = targeting
        self.onAdLoaded = onAdLoaded
        self.onAdFailed = onAdFailed
        self.onAdClicked = onAdClicked
        
        _viewModel = StateObject(wrappedValue: NativeViewModel(
            placementId: placementId,
            targeting: targeting
        ))
    }
    
    public var body: some View {
        Group {
            if viewModel.isLoading {
                ProgressView()
            } else if let ad = viewModel.currentAd {
                nativeAdContent(ad: ad)
            } else {
                Color.clear
            }
        }
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
    private func nativeAdContent(ad: AdResponse) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            // Sponsored badge
            HStack {
                Text("Sponsored")
                    .font(.caption)
                    .foregroundColor(.secondary)
                Spacer()
            }
            
            // Main image
            if let imageURL = ad.imageURL {
                AsyncImage(url: imageURL) { image in
                    image
                        .resizable()
                        .aspectRatio(contentMode: .fill)
                } placeholder: {
                    ProgressView()
                }
                .frame(height: 200)
                .clipped()
                .cornerRadius(8)
            }
            
            // Title
            if let title = ad.title {
                Text(title)
                    .font(.headline)
                    .lineLimit(2)
            }
            
            // Description
            if let description = ad.description {
                Text(description)
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                    .lineLimit(3)
            }
            
            // Call to action button
            if let callToAction = ad.callToAction {
                Button(action: {
                    viewModel.handleClick()
                    onAdClicked?()
                }) {
                    Text(callToAction)
                        .font(.subheadline)
                        .fontWeight(.semibold)
                        .foregroundColor(.white)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.blue)
                        .cornerRadius(8)
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(12)
        .shadow(radius: 2)
        .contentShape(Rectangle())
        .onTapGesture {
            viewModel.handleClick()
            onAdClicked?()
        }
    }
}

// MARK: - Native ViewModel

@MainActor
class NativeViewModel: ObservableObject {
    @Published var currentAd: AdResponse?
    @Published var isLoading = false
    @Published var error: Error?
    
    private let placementId: String
    private let targeting: AdxTargeting?
    
    init(placementId: String, targeting: AdxTargeting?) {
        self.placementId = placementId
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
                    adFormat: .native,
                    adSize: nil,
                    targeting: targeting
                )
                
                // Request ad
                let bidResponse = try await networkClient.requestAd(bidRequest: bidRequest)
                
                // Parse response
                if let ad = BidResponseParser.parse(bidResponse: bidResponse, adFormat: .native) {
                    currentAd = ad
                    
                    // Track impression
                    try? await networkClient.trackImpression(bidId: ad.bidId)
                } else {
                    throw AdxError.noFill
                }
                
                isLoading = false
            } catch {
                self.error = error
                isLoading = false
                AdxSDK.shared.log("Native ad load failed: \(error)", level: .error)
            }
        }
    }
    
    func handleClick() {
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
}
