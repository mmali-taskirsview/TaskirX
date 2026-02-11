package com.taskirx.sdk.internal

import com.taskirx.sdk.AdxSDK
import com.taskirx.sdk.config.AdFormat
import com.taskirx.sdk.config.DeviceType
import com.taskirx.sdk.models.*
import java.util.UUID

/**
 * Builds OpenRTB bid requests
 */
internal class BidRequestBuilder(
    private val deviceInfo: DeviceInfo
) {
    
    suspend fun buildRequest(
        adRequest: AdRequest
    ): BidRequest {
        val impressionId = UUID.randomUUID().toString()
        
        val impression = when (adRequest.adFormat) {
            AdFormat.BANNER, AdFormat.IN_APP_BANNER -> {
                val size = adRequest.adSize ?: throw IllegalArgumentException("Banner ads require adSize")
                Impression(
                    id = impressionId,
                    banner = Banner(
                        width = size.width,
                        height = size.height
                    ),
                    tagId = adRequest.placementId
                )
            }
            AdFormat.VIDEO, AdFormat.IN_APP_VIDEO, AdFormat.REWARDED_VIDEO -> {
                Impression(
                    id = impressionId,
                    video = Video(
                        mimes = listOf("video/mp4", "video/webm"),
                        protocols = listOf(2, 3, 5, 6),
                        width = deviceInfo.getScreenWidth(),
                        height = deviceInfo.getScreenHeight()
                    ),
                    tagId = adRequest.placementId
                )
            }
            AdFormat.NATIVE, AdFormat.IN_APP_NATIVE -> {
                Impression(
                    id = impressionId,
                    native = Native(
                        request = buildNativeRequest()
                    ),
                    tagId = adRequest.placementId
                )
            }
            AdFormat.INTERSTITIAL -> {
                Impression(
                    id = impressionId,
                    banner = Banner(
                        width = deviceInfo.getScreenWidth(),
                        height = deviceInfo.getScreenHeight()
                    ),
                    tagId = adRequest.placementId
                )
            }
        }
        
        val device = Device(
            userAgent = deviceInfo.getUserAgent(),
            deviceType = deviceInfo.getDeviceType().value,
            make = deviceInfo.getManufacturer(),
            model = deviceInfo.getModel(),
            osVersion = deviceInfo.getOsVersion(),
            width = deviceInfo.getScreenWidth(),
            height = deviceInfo.getScreenHeight(),
            advertisingId = deviceInfo.getAdvertisingId(),
            limitAdTracking = if (deviceInfo.isLimitAdTrackingEnabled()) 1 else 0,
            connectionType = deviceInfo.getConnectionType(),
            language = deviceInfo.getLanguage()
        )
        
        val user = User(
            id = AdxSDK.getUserId(),
            gender = adRequest.targeting?.gender?.name?.lowercase(),
            keywords = adRequest.targeting?.keywords?.joinToString(","),
            geo = adRequest.targeting?.location?.let {
                Geo(
                    latitude = it.latitude,
                    longitude = it.longitude,
                    accuracy = it.accuracy?.toInt()
                )
            }
        )
        
        val app = App(
            id = AdxSDK.publisherId,
            name = deviceInfo.getAppName(),
            bundle = deviceInfo.getPackageName(),
            publisher = Publisher(
                id = AdxSDK.publisherId
            )
        )
        
        return BidRequest(
            id = UUID.randomUUID().toString(),
            impressions = listOf(impression),
            device = device,
            user = user,
            app = app,
            test = if (AdxSDK.getConfig().testMode) 1 else 0
        )
    }
    
    private fun buildNativeRequest(): String {
        // Simplified native request
        return """
            {
                "ver": "1.2",
                "assets": [
                    {"id": 1, "required": 1, "title": {"len": 90}},
                    {"id": 2, "required": 0, "img": {"type": 3, "wmin": 300, "hmin": 250}},
                    {"id": 3, "required": 0, "data": {"type": 2, "len": 150}}
                ]
            }
        """.trimIndent()
    }
}
