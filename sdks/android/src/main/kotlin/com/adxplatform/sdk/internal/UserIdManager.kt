package com.taskirx.sdk.internal

import android.content.Context
import android.content.SharedPreferences
import java.util.UUID

/**
 * Manages persistent user ID
 */
internal class UserIdManager(context: Context) {
    
    private val prefs: SharedPreferences = context.getSharedPreferences(
        "adx_sdk_prefs",
        Context.MODE_PRIVATE
    )
    
    companion object {
        private const val KEY_USER_ID = "user_id"
        private const val KEY_CUSTOM_USER_ID = "custom_user_id"
    }
    
    /**
     * Get or generate persistent user ID
     */
    fun getUserId(): String {
        // Check for custom user ID first
        val customId = prefs.getString(KEY_CUSTOM_USER_ID, null)
        if (!customId.isNullOrBlank()) {
            return customId
        }
        
        // Get or generate SDK user ID
        var userId = prefs.getString(KEY_USER_ID, null)
        if (userId.isNullOrBlank()) {
            userId = UUID.randomUUID().toString()
            prefs.edit().putString(KEY_USER_ID, userId).apply()
        }
        return userId
    }
    
    /**
     * Set custom user ID (from app)
     */
    fun setCustomUserId(userId: String) {
        prefs.edit().putString(KEY_CUSTOM_USER_ID, userId).apply()
    }
    
    /**
     * Clear custom user ID
     */
    fun clearCustomUserId() {
        prefs.edit().remove(KEY_CUSTOM_USER_ID).apply()
    }
}
