package com.taskirx.sdk.ads

import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.Image
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import coil.compose.rememberAsyncImagePainter
import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.config.AdFormat
import com.taskirx.sdk.config.AdSize
import com.taskirx.sdk.internal.BidRequestBuilder
import com.taskirx.sdk.internal.BidResponseParser
import com.taskirx.sdk.models.AdError
import com.taskirx.sdk.models.AdRequest
import com.taskirx.sdk.models.AdResponse
import com.taskirx.sdk.models.ErrorCode
import kotlinx.coroutines.launch

/**
 * Jetpack Compose banner ad component
 * 
 * Usage:
 * ```
 * AdxBannerView(
 *     placementId = "banner-home",
 *     adSize = AdSize.BANNER_320x50,
 *     onAdLoaded = { Log.d("Ad", "Loaded") },
 *     onAdFailed = { error -> Log.e("Ad", error.message) }
 * )
 * ```
 */
@Composable
fun AdxBannerView(
    placementId: String,
    adSize: AdSize,
    modifier: Modifier = Modifier,
    onAdLoaded: (() -> Unit)? = null,
    onAdFailed: ((AdError) -> Unit)? = null,
    onAdClicked: (() -> Unit)? = null,
    onAdImpression: (() -> Unit)? = null
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    
    var adResponse by remember { mutableStateOf<AdResponse?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<AdError?>(null) }
    var impressionTracked by remember { mutableStateOf(false) }
    
    // Load ad on composition
    LaunchedEffect(placementId) {
        scope.launch {
            try {
                val adRequest = AdRequest(
                    placementId = placementId,
                    adFormat = AdFormat.BANNER,
                    adSize = adSize
                )
                
                val bidRequest = BidRequestBuilder(AdxSDK.getDeviceInfo())
                    .buildRequest(adRequest)
                
                val result = AdxSDK.getNetworkClient().requestAd(bidRequest)
                
                result.onSuccess { bidResponse ->
                    val ad = BidResponseParser.parse(bidResponse, AdFormat.BANNER)
                    if (ad != null) {
                        adResponse = ad
                        isLoading = false
                        onAdLoaded?.invoke()
                    } else {
                        val noFillError = AdError(
                            code = ErrorCode.NO_FILL,
                            message = "No ad available"
                        )
                        error = noFillError
                        isLoading = false
                        onAdFailed?.invoke(noFillError)
                    }
                }.onFailure { e ->
                    val loadError = AdError(
                        code = ErrorCode.AD_LOAD_FAILED,
                        message = e.message ?: "Failed to load ad",
                        cause = e
                    )
                    error = loadError
                    isLoading = false
                    onAdFailed?.invoke(loadError)
                }
            } catch (e: Exception) {
                val loadError = AdError(
                    code = ErrorCode.INTERNAL_ERROR,
                    message = e.message ?: "Internal error",
                    cause = e
                )
                error = loadError
                isLoading = false
                onAdFailed?.invoke(loadError)
            }
        }
    }
    
    // Track impression when ad is visible
    LaunchedEffect(adResponse) {
        if (adResponse != null && !impressionTracked) {
            scope.launch {
                AdxSDK.getNetworkClient().trackImpression(adResponse!!.bidId)
                impressionTracked = true
                onAdImpression?.invoke()
            }
        }
    }
    
    Box(
        modifier = modifier
            .width(adSize.width.dp)
            .height(adSize.height.dp),
        contentAlignment = Alignment.Center
    ) {
        when {
            isLoading -> {
                CircularProgressIndicator()
            }
            error != null -> {
                Text("Ad failed to load")
            }
            adResponse != null -> {
                BannerAdContent(
                    adResponse = adResponse!!,
                    context = context,
                    onAdClicked = {
                        scope.launch {
                            AdxSDK.getNetworkClient().trackClick(adResponse!!.bidId)
                            onAdClicked?.invoke()
                        }
                    }
                )
            }
        }
    }
}

@Composable
private fun BannerAdContent(
    adResponse: AdResponse,
    context: Context,
    onAdClicked: () -> Unit
) {
    val painter = rememberAsyncImagePainter(adResponse.imageUrl)
    
    Image(
        painter = painter,
        contentDescription = "Banner ad",
        contentScale = ContentScale.Fit,
        modifier = Modifier
            .fillMaxSize()
            .clickable {
                onAdClicked()
                adResponse.clickUrl?.let { url ->
                    try {
                        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
                        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                        context.startActivity(intent)
                    } catch (e: Exception) {
                        AdxSDK.logError("Failed to open click URL", e)
                    }
                }
            }
    )
}

/**
 * Traditional View-based banner ad (for non-Compose apps)
 */
class AdxBannerViewLegacy(context: Context) : androidx.compose.ui.platform.ComposeView(context) {
    
    private var placementId: String = ""
    private var adSize: AdSize = AdSize.BANNER_320x50
    private var onAdLoaded: (() -> Unit)? = null
    private var onAdFailed: ((AdError) -> Unit)? = null
    private var onAdClicked: (() -> Unit)? = null
    private var onAdImpression: (() -> Unit)? = null
    
    /**
     * Load banner ad
     */
    fun load(
        placementId: String,
        adSize: AdSize,
        onAdLoaded: (() -> Unit)? = null,
        onAdFailed: ((AdError) -> Unit)? = null,
        onAdClicked: (() -> Unit)? = null,
        onAdImpression: (() -> Unit)? = null
    ) {
        this.placementId = placementId
        this.adSize = adSize
        this.onAdLoaded = onAdLoaded
        this.onAdFailed = onAdFailed
        this.onAdClicked = onAdClicked
        this.onAdImpression = onAdImpression
        
        setContent {
            AdxBannerView(
                placementId = placementId,
                adSize = adSize,
                onAdLoaded = onAdLoaded,
                onAdFailed = onAdFailed,
                onAdClicked = onAdClicked,
                onAdImpression = onAdImpression
            )
        }
    }
}
