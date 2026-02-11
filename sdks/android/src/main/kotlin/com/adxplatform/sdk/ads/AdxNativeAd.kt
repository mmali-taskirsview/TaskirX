package com.taskirx.sdk.ads

import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.Image
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import coil.compose.rememberAsyncImagePainter
import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.models.AdResponse

/**
 * Native ad data holder
 */
data class AdxNativeAdData(
    val title: String,
    val description: String,
    val imageUrl: String?,
    val iconUrl: String?,
    val callToAction: String,
    val sponsoredBy: String?,
    val clickUrl: String?,
    internal val bidId: String,
    internal val context: Context
) {
    /**
     * Call this when user clicks on the ad
     */
    fun performClick() {
        // Track click
        kotlinx.coroutines.GlobalScope.launch {
            AdxSDK.getNetworkClient().trackClick(bidId)
        }
        
        // Open URL
        clickUrl?.let { url ->
            try {
                val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
                intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                context.startActivity(intent)
            } catch (e: Exception) {
                AdxSDK.logError("Failed to open click URL", e)
            }
        }
    }
}

/**
 * Pre-built native ad template
 */
@Composable
fun AdxNativeAdTemplate(
    nativeAd: AdxNativeAdData,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier
            .fillMaxWidth()
            .padding(8.dp)
            .clickable { nativeAd.performClick() }
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            // Main image
            nativeAd.imageUrl?.let { imageUrl ->
                Image(
                    painter = rememberAsyncImagePainter(imageUrl),
                    contentDescription = "Ad image",
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(200.dp)
                )
                Spacer(modifier = Modifier.height(12.dp))
            }
            
            // Icon and sponsored label
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.fillMaxWidth()
            ) {
                nativeAd.iconUrl?.let { iconUrl ->
                    Image(
                        painter = rememberAsyncImagePainter(iconUrl),
                        contentDescription = "Icon",
                        modifier = Modifier.size(32.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                }
                
                Column {
                    nativeAd.sponsoredBy?.let { sponsor ->
                        Text(
                            text = sponsor,
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                    Text(
                        text = "Sponsored",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            // Title
            Text(
                text = nativeAd.title,
                style = MaterialTheme.typography.titleMedium
            )
            
            Spacer(modifier = Modifier.height(4.dp))
            
            // Description
            Text(
                text = nativeAd.description,
                style = MaterialTheme.typography.bodyMedium
            )
            
            Spacer(modifier = Modifier.height(12.dp))
            
            // Call to action button
            Button(
                onClick = { nativeAd.performClick() },
                modifier = Modifier.align(Alignment.End)
            ) {
                Text(nativeAd.callToAction)
            }
        }
    }
}

/**
 * Helper to convert AdResponse to AdxNativeAdData
 */
internal fun AdResponse.toNativeAdData(context: Context): AdxNativeAdData {
    return AdxNativeAdData(
        title = title ?: "Untitled",
        description = description ?: "",
        imageUrl = imageUrl,
        iconUrl = iconUrl,
        callToAction = callToAction ?: "Learn More",
        sponsoredBy = sponsoredBy,
        clickUrl = clickUrl,
        bidId = bidId,
        context = context
    )
}
