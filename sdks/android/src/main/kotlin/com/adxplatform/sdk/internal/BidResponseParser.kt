package com.taskirx.sdk.internal

import com.taskirx.sdk.config.AdFormat
import com.taskirx.sdk.models.AdResponse
import com.taskirx.sdk.models.BidResponse
import org.json.JSONObject

/**
 * Parses OpenRTB bid responses into AdResponse objects
 */
internal object BidResponseParser {
    
    fun parse(bidResponse: BidResponse, adFormat: AdFormat): AdResponse? {
        val seatBid = bidResponse.seatBids?.firstOrNull() ?: return null
        val bid = seatBid.bids.firstOrNull() ?: return null
        
        return when (adFormat) {
            AdFormat.BANNER, AdFormat.IN_APP_BANNER -> parseBanner(bid, bidResponse.bidId ?: bid.id)
            AdFormat.VIDEO, AdFormat.IN_APP_VIDEO, AdFormat.REWARDED_VIDEO -> parseVideo(bid, bidResponse.bidId ?: bid.id)
            AdFormat.NATIVE, AdFormat.IN_APP_NATIVE -> parseNative(bid, bidResponse.bidId ?: bid.id)
            AdFormat.INTERSTITIAL -> parseBanner(bid, bidResponse.bidId ?: bid.id)
        }
    }
    
    private fun parseBanner(bid: com.taskirx.sdk.models.Bid, bidId: String): AdResponse {
        // Ad markup should contain image URL or HTML
        val markup = bid.adMarkup
        
        // Try to extract image URL from markup
        val imageUrl = extractImageUrl(markup) ?: markup
        val clickUrl = extractClickUrl(markup)
        
        return AdResponse(
            bidId = bidId,
            impressionId = bid.impressionId,
            price = bid.price,
            adFormat = AdFormat.BANNER,
            imageUrl = imageUrl,
            clickUrl = clickUrl,
            impressionUrl = bid.noticeUrl,
            width = bid.width,
            height = bid.height,
            extensions = bid.extensions
        )
    }
    
    private fun parseVideo(bid: com.taskirx.sdk.models.Bid, bidId: String): AdResponse {
        // Ad markup should contain VAST XML or video URL
        val videoUrl = extractVideoUrl(bid.adMarkup) ?: bid.adMarkup
        val clickUrl = extractClickUrl(bid.adMarkup)
        
        return AdResponse(
            bidId = bidId,
            impressionId = bid.impressionId,
            price = bid.price,
            adFormat = AdFormat.VIDEO,
            videoUrl = videoUrl,
            clickUrl = clickUrl,
            impressionUrl = bid.noticeUrl,
            extensions = bid.extensions
        )
    }
    
    private fun parseNative(bid: com.taskirx.sdk.models.Bid, bidId: String): AdResponse {
        // Parse native ad markup (JSON)
        try {
            val json = JSONObject(bid.adMarkup)
            val native = json.optJSONObject("native")
            
            var title: String? = null
            var description: String? = null
            var imageUrl: String? = null
            var iconUrl: String? = null
            var clickUrl: String? = null
            var callToAction: String? = null
            var sponsoredBy: String? = null
            
            // Parse assets
            val assets = native?.optJSONArray("assets")
            if (assets != null) {
                for (i in 0 until assets.length()) {
                    val asset = assets.getJSONObject(i)
                    val id = asset.optInt("id")
                    
                    when (id) {
                        1 -> title = asset.optJSONObject("title")?.optString("text")
                        2 -> imageUrl = asset.optJSONObject("img")?.optString("url")
                        3 -> description = asset.optJSONObject("data")?.optString("value")
                    }
                }
            }
            
            // Parse link
            val link = native?.optJSONObject("link")
            clickUrl = link?.optString("url")
            
            return AdResponse(
                bidId = bidId,
                impressionId = bid.impressionId,
                price = bid.price,
                adFormat = AdFormat.NATIVE,
                title = title,
                description = description,
                imageUrl = imageUrl,
                iconUrl = iconUrl,
                clickUrl = clickUrl,
                callToAction = callToAction ?: "Learn More",
                sponsoredBy = sponsoredBy,
                impressionUrl = bid.noticeUrl,
                extensions = bid.extensions
            )
        } catch (e: Exception) {
            throw IllegalArgumentException("Failed to parse native ad response", e)
        }
    }
    
    private fun extractImageUrl(markup: String): String? {
        // Extract image URL from HTML
        val imgRegex = """<img[^>]+src=["']([^"']+)["']""".toRegex()
        return imgRegex.find(markup)?.groupValues?.get(1)
    }
    
    private fun extractClickUrl(markup: String): String? {
        // Extract click URL from HTML
        val hrefRegex = """<a[^>]+href=["']([^"']+)["']""".toRegex()
        return hrefRegex.find(markup)?.groupValues?.get(1)
    }
    
    private fun extractVideoUrl(markup: String): String? {
        // Extract video URL from VAST XML or markup
        val videoRegex = """<MediaFile[^>]*>([^<]+)</MediaFile>""".toRegex()
        val match = videoRegex.find(markup)?.groupValues?.get(1)?.trim()
        
        // If not VAST, check for direct video URL
        return match ?: if (markup.contains(".mp4") || markup.contains(".webm")) {
            markup.trim()
        } else {
            null
        }
    }
}
