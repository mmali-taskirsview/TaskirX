package com.taskirx.sdk.ui

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.Image
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import coil.compose.rememberAsyncImagePainter
import com.taskirx.sdk.AdxSDK
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/**
 * Full-screen interstitial ad activity
 */
class InterstitialAdActivity : ComponentActivity() {
    
    private var bidId: String? = null
    private var imageUrl: String? = null
    private var clickUrl: String? = null
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        bidId = intent.getStringExtra("bidId")
        imageUrl = intent.getStringExtra("imageUrl")
        clickUrl = intent.getStringExtra("clickUrl")
        
        setContent {
            MaterialTheme {
                InterstitialAdScreen(
                    imageUrl = imageUrl,
                    onClose = { finish() },
                    onAdClick = {
                        // Track click
                        kotlinx.coroutines.GlobalScope.launch {
                            bidId?.let { AdxSDK.getNetworkClient().trackClick(it) }
                        }
                        
                        // Open URL
                        clickUrl?.let { url ->
                            try {
                                val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
                                startActivity(intent)
                            } catch (e: Exception) {
                                AdxSDK.logError("Failed to open URL", e)
                            }
                        }
                    }
                )
            }
        }
    }
}

@Composable
private fun InterstitialAdScreen(
    imageUrl: String?,
    onClose: () -> Unit,
    onAdClick: () -> Unit
) {
    var showCloseButton by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()
    
    // Show close button after 5 seconds
    LaunchedEffect(Unit) {
        delay(5000)
        showCloseButton = true
    }
    
    Box(
        modifier = Modifier.fillMaxSize()
    ) {
        // Ad image (clickable)
        imageUrl?.let { url ->
            Image(
                painter = rememberAsyncImagePainter(url),
                contentDescription = "Interstitial ad",
                contentScale = ContentScale.Fit,
                modifier = Modifier
                    .fillMaxSize()
                    .clickable { onAdClick() }
            )
        }
        
        // Close button (top-right corner)
        if (showCloseButton) {
            IconButton(
                onClick = onClose,
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(16.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Close,
                    contentDescription = "Close ad",
                    tint = MaterialTheme.colorScheme.onSurface
                )
            }
        }
    }
}
