package com.taskirx.sdk.models

import com.taskirx.sdk.config.AdFormat
import com.taskirx.sdk.config.AdSize
import com.taskirx.sdk.config.Gender
import com.squareup.moshi.Json
import com.squareup.moshi.JsonClass

/**
 * Ad request parameters
 */
data class AdRequest(
    val placementId: String,
    val adFormat: AdFormat,
    val adSize: AdSize? = null,
    val targeting: AdTargeting? = null,
    val customData: Map<String, String>? = null
)

/**
 * Targeting parameters
 */
data class AdTargeting(
    val age: Int? = null,
    val gender: Gender? = null,
    val interests: List<String>? = null,
    val keywords: List<String>? = null,
    val location: Location? = null,
    val customTargeting: Map<String, String>? = null
)

/**
 * Location data
 */
data class Location(
    val latitude: Double,
    val longitude: Double,
    val accuracy: Float? = null
)

/**
 * OpenRTB Bid Request
 */
@JsonClass(generateAdapter = true)
data class BidRequest(
    @Json(name = "id") val id: String,
    @Json(name = "imp") val impressions: List<Impression>,
    @Json(name = "device") val device: Device,
    @Json(name = "user") val user: User,
    @Json(name = "app") val app: App,
    @Json(name = "test") val test: Int = 0
)

@JsonClass(generateAdapter = true)
data class Impression(
    @Json(name = "id") val id: String,
    @Json(name = "banner") val banner: Banner? = null,
    @Json(name = "video") val video: Video? = null,
    @Json(name = "native") val native: Native? = null,
    @Json(name = "bidfloor") val bidFloor: Double = 0.0,
    @Json(name = "tagid") val tagId: String
)

@JsonClass(generateAdapter = true)
data class Banner(
    @Json(name = "w") val width: Int,
    @Json(name = "h") val height: Int,
    @Json(name = "format") val formats: List<Format>? = null
)

@JsonClass(generateAdapter = true)
data class Format(
    @Json(name = "w") val width: Int,
    @Json(name = "h") val height: Int
)

@JsonClass(generateAdapter = true)
data class Video(
    @Json(name = "mimes") val mimes: List<String>,
    @Json(name = "minduration") val minDuration: Int = 5,
    @Json(name = "maxduration") val maxDuration: Int = 30,
    @Json(name = "protocols") val protocols: List<Int>,
    @Json(name = "w") val width: Int,
    @Json(name = "h") val height: Int
)

@JsonClass(generateAdapter = true)
data class Native(
    @Json(name = "request") val request: String
)

@JsonClass(generateAdapter = true)
data class Device(
    @Json(name = "ua") val userAgent: String,
    @Json(name = "ip") val ip: String = "",
    @Json(name = "devicetype") val deviceType: Int,
    @Json(name = "make") val make: String,
    @Json(name = "model") val model: String,
    @Json(name = "os") val os: String = "Android",
    @Json(name = "osv") val osVersion: String,
    @Json(name = "w") val width: Int,
    @Json(name = "h") val height: Int,
    @Json(name = "ifa") val advertisingId: String? = null,
    @Json(name = "lmt") val limitAdTracking: Int = 0,
    @Json(name = "connectiontype") val connectionType: Int = 0,
    @Json(name = "language") val language: String
)

@JsonClass(generateAdapter = true)
data class User(
    @Json(name = "id") val id: String,
    @Json(name = "yob") val yearOfBirth: Int? = null,
    @Json(name = "gender") val gender: String? = null,
    @Json(name = "keywords") val keywords: String? = null,
    @Json(name = "geo") val geo: Geo? = null
)

@JsonClass(generateAdapter = true)
data class Geo(
    @Json(name = "lat") val latitude: Double,
    @Json(name = "lon") val longitude: Double,
    @Json(name = "type") val type: Int = 1,
    @Json(name = "accuracy") val accuracy: Int? = null,
    @Json(name = "country") val country: String? = null,
    @Json(name = "city") val city: String? = null
)

@JsonClass(generateAdapter = true)
data class App(
    @Json(name = "id") val id: String,
    @Json(name = "name") val name: String,
    @Json(name = "bundle") val bundle: String,
    @Json(name = "storeurl") val storeUrl: String? = null,
    @Json(name = "publisher") val publisher: Publisher
)

@JsonClass(generateAdapter = true)
data class Publisher(
    @Json(name = "id") val id: String,
    @Json(name = "name") val name: String? = null
)

/**
 * OpenRTB Bid Response
 */
@JsonClass(generateAdapter = true)
data class BidResponse(
    @Json(name = "id") val id: String,
    @Json(name = "seatbid") val seatBids: List<SeatBid>?,
    @Json(name = "bidid") val bidId: String?,
    @Json(name = "cur") val currency: String = "USD",
    @Json(name = "nbr") val noBidReason: Int? = null
)

@JsonClass(generateAdapter = true)
data class SeatBid(
    @Json(name = "bid") val bids: List<Bid>,
    @Json(name = "seat") val seat: String? = null
)

@JsonClass(generateAdapter = true)
data class Bid(
    @Json(name = "id") val id: String,
    @Json(name = "impid") val impressionId: String,
    @Json(name = "price") val price: Double,
    @Json(name = "adid") val adId: String? = null,
    @Json(name = "nurl") val noticeUrl: String? = null,
    @Json(name = "adm") val adMarkup: String,
    @Json(name = "adomain") val advertiserDomains: List<String>? = null,
    @Json(name = "crid") val creativeId: String? = null,
    @Json(name = "w") val width: Int? = null,
    @Json(name = "h") val height: Int? = null,
    @Json(name = "ext") val extensions: Map<String, Any>? = null
)

/**
 * Parsed ad response for SDK use
 */
data class AdResponse(
    val bidId: String,
    val impressionId: String,
    val price: Double,
    val adFormat: AdFormat,
    val imageUrl: String? = null,
    val clickUrl: String? = null,
    val impressionUrl: String? = null,
    val videoUrl: String? = null,
    val title: String? = null,
    val description: String? = null,
    val callToAction: String? = null,
    val iconUrl: String? = null,
    val sponsoredBy: String? = null,
    val width: Int? = null,
    val height: Int? = null,
    val extensions: Map<String, Any>? = null
)

/**
 * Ad error
 */
data class AdError(
    val code: ErrorCode,
    val message: String,
    val cause: Throwable? = null
)

enum class ErrorCode {
    NOT_INITIALIZED,
    NETWORK_ERROR,
    NO_FILL,
    INVALID_REQUEST,
    INTERNAL_ERROR,
    TIMEOUT,
    AD_LOAD_FAILED,
    AD_SHOW_FAILED,
    INVALID_PLACEMENT
}
