package com.taskirx.sdk.ads

import android.content.Context
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
 * Reward information
 */
data class AdxReward(
    val type: String,
    val amount: Int
)

/**
 * Rewarded video ad
 * 
 * Usage:
 * ```
 * val rewardedAd = AdxRewardedVideo(context)
 * rewardedAd.load(
 *     placementId = "rewarded-extra-lives",
 *     onAdLoaded = { rewardedAd.show() },
 *     onRewarded = { reward ->
 *         // Give user reward
 *         giveUserCoins(reward.amount)
 *     }
 * )
 * ```
 */
class AdxRewardedVideo(private val context: Context) {
    
    private var adResponse: AdResponse? = null
    private var isLoaded = false
    
    private var onAdLoaded: (() -> Unit)? = null
    private var onAdFailed: ((AdError) -> Unit)? = null
    private var onAdShown: (() -> Unit)? = null
    private var onRewarded: ((AdxReward) -> Unit)? = null
    private var onAdClosed: (() -> Unit)? = null
    
    /**
     * Load rewarded video ad
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
                    adFormat = AdFormat.REWARDED_VIDEO,
                    adSize = AdSize.BANNER_320x50
                )
                
                val bidRequest = BidRequestBuilder(AdxSDK.getDeviceInfo())
                    .buildRequest(adRequest)
                
                val result = withContext(Dispatchers.IO) {
                    AdxSDK.getNetworkClient().requestAd(bidRequest)
                }
                
                result.onSuccess { bidResponse ->
                    val ad = BidResponseParser.parse(bidResponse, AdFormat.REWARDED_VIDEO)
                    if (ad != null) {
                        adResponse = ad
                        isLoaded = true
                        onAdLoaded?.invoke()
                        AdxSDK.log("Rewarded video ad loaded: $placementId")
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
     * Show rewarded video ad (must be loaded first)
     */
    fun show(
        onAdShown: (() -> Unit)? = null,
        onRewarded: ((AdxReward) -> Unit)? = null,
        onAdClosed: (() -> Unit)? = null
    ) {
        this.onAdShown = onAdShown
        this.onRewarded = onRewarded
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
        val intent = android.content.Intent(context, com.taskirx.sdk.ui.RewardedVideoActivity::class.java)
        intent.putExtra("bidId", adResponse!!.bidId)
        intent.putExtra("videoUrl", adResponse!!.videoUrl)
        intent.putExtra("clickUrl", adResponse!!.clickUrl)
        intent.addFlags(android.content.Intent.FLAG_ACTIVITY_NEW_TASK)
        
        // Store callbacks for activity to access
        rewardedVideoCallbacks = RewardedVideoCallbacks(
            onRewarded = onRewarded,
            onAdClosed = onAdClosed
        )
        
        context.startActivity(intent)
        
        onAdShown?.invoke()
        AdxSDK.log("Rewarded video ad shown")
        
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
        onRewarded = null
        onAdClosed = null
    }
    
    companion object {
        internal var rewardedVideoCallbacks: RewardedVideoCallbacks? = null
    }
}

/**
 * Internal callbacks holder
 */
internal data class RewardedVideoCallbacks(
    val onRewarded: ((AdxReward) -> Unit)?,
    val onAdClosed: (() -> Unit)?
)
