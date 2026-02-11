package com.taskirx.sdk.ads

import android.content.Context
import androidx.lifecycle.lifecycleScope
import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.config.AdFormat
import com.taskirx.sdk.config.AdSize
import com.taskirx.sdk.internal.BidRequestBuilder
import com.taskirx.sdk.internal.BidResponseParser
import com.taskirx.sdk.models.AdError
import com.taskirx.sdk.models.AdRequest
import com.taskirx.sdk.models.AdResponse
import com.taskirx.sdk.models.ErrorCode
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

/**
 * Interstitial ad (full-screen)
 * 
 * Usage:
 * ```
 * val interstitial = AdxInterstitial(context)
 * interstitial.load(
 *     placementId = "interstitial-level-complete",
 *     onAdLoaded = { interstitial.show() },
 *     onAdFailed = { error -> Log.e("Ad", error.message) }
 * )
 * ```
 */
class AdxInterstitial(private val context: Context) {
    
    private var adResponse: AdResponse? = null
    private var isLoaded = false
    
    private var onAdLoaded: (() -> Unit)? = null
    private var onAdFailed: ((AdError) -> Unit)? = null
    private var onAdShown: (() -> Unit)? = null
    private var onAdClicked: (() -> Unit)? = null
    private var onAdClosed: (() -> Unit)? = null
    
    /**
     * Load interstitial ad
     */
    fun load(
        placementId: String,
        onAdLoaded: (() -> Unit)? = null,
        onAdFailed: ((AdError) -> Unit)? = null
    ) {
        this.onAdLoaded = onAdLoaded
        this.onAdFailed = onAdFailed
        
        CoroutineScope(Dispatchers.Main).launch {
            try {
                val adRequest = AdRequest(
                    placementId = placementId,
                    adFormat = AdFormat.INTERSTITIAL,
                    adSize = AdSize.BANNER_320x50 // Full screen
                )
                
                val bidRequest = BidRequestBuilder(AdxSDK.getDeviceInfo())
                    .buildRequest(adRequest)
                
                val result = withContext(Dispatchers.IO) {
                    AdxSDK.getNetworkClient().requestAd(bidRequest)
                }
                
                result.onSuccess { bidResponse ->
                    val ad = BidResponseParser.parse(bidResponse, AdFormat.INTERSTITIAL)
                    if (ad != null) {
                        adResponse = ad
                        isLoaded = true
                        onAdLoaded?.invoke()
                        AdxSDK.log("Interstitial ad loaded: $placementId")
                    } else {
                        val error = AdError(
                            code = ErrorCode.NO_FILL,
                            message = "No ad available"
                        )
                        onAdFailed?.invoke(error)
                    }
                }.onFailure { e ->
                    val error = AdError(
                        code = ErrorCode.AD_LOAD_FAILED,
                        message = e.message ?: "Failed to load ad",
                        cause = e
                    )
                    onAdFailed?.invoke(error)
                }
            } catch (e: Exception) {
                val error = AdError(
                    code = ErrorCode.INTERNAL_ERROR,
                    message = e.message ?: "Internal error",
                    cause = e
                )
                onAdFailed?.invoke(error)
            }
        }
    }
    
    /**
     * Show interstitial ad (must be loaded first)
     */
    fun show(
        onAdShown: (() -> Unit)? = null,
        onAdClicked: (() -> Unit)? = null,
        onAdClosed: (() -> Unit)? = null
    ) {
        this.onAdShown = onAdShown
        this.onAdClicked = onAdClicked
        this.onAdClosed = onAdClosed
        
        if (!isLoaded || adResponse == null) {
            val error = AdError(
                code = ErrorCode.AD_SHOW_FAILED,
                message = "Ad not loaded. Call load() first."
            )
            onAdFailed?.invoke(error)
            return
        }
        
        // Track impression
        CoroutineScope(Dispatchers.IO).launch {
            AdxSDK.getNetworkClient().trackImpression(adResponse!!.bidId)
        }
        
        // Show ad in activity
        val intent = android.content.Intent(context, com.taskirx.sdk.ui.InterstitialAdActivity::class.java)
        intent.putExtra("bidId", adResponse!!.bidId)
        intent.putExtra("imageUrl", adResponse!!.imageUrl)
        intent.putExtra("clickUrl", adResponse!!.clickUrl)
        intent.addFlags(android.content.Intent.FLAG_ACTIVITY_NEW_TASK)
        
        context.startActivity(intent)
        
        onAdShown?.invoke()
        AdxSDK.log("Interstitial ad shown")
        
        // Reset state
        isLoaded = false
        adResponse = null
    }
    
    /**
     * Check if ad is loaded and ready to show
     */
    fun isReady(): Boolean = isLoaded && adResponse != null
    
    /**
     * Destroy and release resources
     */
    fun destroy() {
        adResponse = null
        isLoaded = false
        onAdLoaded = null
        onAdFailed = null
        onAdShown = null
        onAdClicked = null
        onAdClosed = null
    }
}
